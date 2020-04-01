# 作業手順書

## 最初のdeploy
- vpc, security groupを作る
- clusterを一つ立ち上げる
- ELB(Application)を立ち上げる。target groupは2つ
- ECRにfrontとapiの2つのコンテナを登録してpush
- fargate serviceを2つ作る

### 前提条件
- ローカルPCにインストールしておくもの
  - [aws cli version2](https://docs.aws.amazon.com/ja_jp/cli/latest/userguide/install-cliv2-linux.html)
  - [`aws configure`](https://docs.aws.amazon.com/ja_jp/cli/latest/userguide/cli-chap-configure.html#cli-quick-configuration) IAMは [Creating Your First IAM Admin User and Group](https://docs.aws.amazon.com/IAM/latest/UserGuide/getting-started_create-admin-group.html) に従いフル権限を持つものを使う。入力するAWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEYはIAMのコンソールから、default region nameは`ap-northeast-1`を使う。
  - [ecs cli](https://docs.aws.amazon.com/ja_jp/AmazonECS/latest/developerguide/ECS_CLI_installation.html)
  - ecs cliのconfigureは後で行う
### 手順
- 以下のコマンドを打っていく
```
$ git clone https://github.com/uta8a/goping.git
$ cd goping
```
- cloudformationでvpcとsecurity groupを作る
```
$ aws cloudformation create-stack --stack-name fargate-sample-vpc --capabilities CAPABILITY_IAM --template-body file://.cloudformation/vpc.yml
$ aws cloudformation create-stack --stack-name fargate-sample-sg --template-body file://.cloudformation/security_group.yml
```
- ここで、vpc作成後にすぐにsecurity_groupを作ろうとすると失敗してsecurity groupがrollbackすることがあるので、AWS consoleからcloudformationを見に行ってきちんと両方作成されているか確認する。
- もしsecurity groupの作成に失敗していたら、security groupのcloudformationをdeleteしてもう一度コマンドを打つ
- 次に環境変数の設定と`ecs-cli configure`を行う
```
$ export AWS_ACCESS_KEY_ID="XXX" AWS_SECRET_ACCESS_KEY="XXX"
```
- この値は`aws configure`で取得したもの
```
$ ecs-cli configure --cluster fargate-sample-cluster --region ap-northeast-1 --default-launch-type FARGATE --config-name fargate-sample
$ ecs-cli configure profile default --profile-name fargate-sample
```
- ここでcloudformation > fargate-sample-vpc > Resourcesを見てsubnet 2つとvpcのidをメモしておく
```
$ export VPC="XXX" SUBNET_1="XXX" SUBNET_2="XXX"
```
- 次に、ecs-cli upでfargateを指定しclusterを立ち上げる
```
$ ecs-cli up --cluster fargate-sample-cluster --vpc $VPC --subnets $SUBNET_1,$SUBNET_2  --security-group fargate-sample-sg --launch-type FARGATE --region ap-northeast-1 --ecs-profile fargate-sample
```
- 次に、Load Balancerの設定と紐付けを行う
```
$ aws cloudformation create-stack --stack-name fargate-sample-alb --template-body file://.cloudformation/alb.yml
```
- cloudformationのコンソールでALBができるのを待つ。(5分ほど時間がかかる)
- ここで`goping-dev-alb`という名前のALBが作成されるのでDNS nameを控えておく
- DNS nameはfrontからapiにアクセスするときに使うので、VUE_APP_API_HOSTに設定しておく
```
$ export VUE_APP_API_HOST="XXX"
```
- 次に、docker imageをECRに登録してpushする。
- まずはfrontのコンテナから
- AWS_ACCOUNT_IDはMy Accountから見れる
```
$ export AWS_ACCOUNT_ID="XXX" AWS_DEFAULT_REGION="ap-northeast-1"
$ aws ecr get-login-password --region $AWS_DEFAULT_REGION | docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com
$ docker build -t goping-front www/
$ aws ecr create-repository \
    --repository-name goping-front \
    --image-scanning-configuration scanOnPush=true \
    --region $AWS_DEFAULT_REGION
$ docker tag goping-front:latest $AWS_ACCOUNT_ID.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com/goping-front:latest
$ docker push $AWS_ACCOUNT_ID.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com/goping-front:latest
```
- 続けてapiのコンテナ
```
$ docker build -t goping-api api/
$ aws ecr create-repository \
    --repository-name goping-api \
    --image-scanning-configuration scanOnPush=true \
    --region $AWS_DEFAULT_REGION
$ docker tag goping-api:latest $AWS_ACCOUNT_ID.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com/goping-api:latest
$ docker push $AWS_ACCOUNT_ID.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com/goping-api:latest
```
- 次に、serviceの立ち上げをする
- SECURITY_GROUPはEC2 > security group > fargate-sample-sgのidから見れる
```
$ export SECURITY_GROUP="XXX"
```
- ALBでtarget groupを2つ作っておいたのでそれぞれを設定しておく。TargetGroupApiはAPI側に、TargetGroupFrontはfront側に設定
```
$ export TG_API="XXX" TG_FRONT="XXX"
```
- `ecs-cli compose service up`でサービスを立ち上げる
- まずapiのサービスから
```
$ ecs-cli compose --project-name goping-api -f docker-compose.ecs.yml --ecs-params ecs-params.yml --cluster fargate-sample-cluster service up --deployment-max-percent 200 --deployment-min-healthy-percent 50 --target-group-arn $TG_API --container-name api --container-port 8001 --launch-type FARGATE --health-check-grace-period 120 --create-log-groups --timeout 10
```
- 次にfrontのサービス
```
$ ecs-cli compose --project-name goping-front -f docker-compose.ecs.yml --ecs-params ecs-params.yml --cluster fargate-sample-cluster service up --deployment-max-percent 200 --deployment-min-healthy-percent 50 --target-group-arn $TG_FRONT --container-name frontend --container-port 80 --launch-type FARGATE --health-check-grace-period 120 --create-log-groups --timeout 10
```
- これで一通り完成。ECSからfargateの様子を見て、大丈夫そうならALBのDNS nameからアクセスすると見れる。
- うまく行かない場合はCloudWatchからLogを見るとよい。

## CI
- GitHubリポジトリとCircleCIを紐付ける
- CircleCI側に環境変数を設定する
- 設定するものは以下の通り
```
AWS_ACCESS_KEY_ID
AWS_ACCOUNT_ID
AWS_DEFAULT_REGION
AWS_ECR_ACCOUNT_URL
AWS_SECRET_ACCESS_KEY
SECURITY_GROUP
SUBNET_1
SUBNET_2
VUE_APP_API_HOST
```
- AWS_ECR_ACCOUNT_URLは`$AWS_ACCOUNT_ID.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com`を設定すればよい
- あとはGitHubにpushすればCircleCIが走る
- DNS nameにアクセスしてうまくいっているか確認
- 変だと思ったらCloudWatchを確認