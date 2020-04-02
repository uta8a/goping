# 作業手順書

- [最初のdeploy][title-1]
- [CI][title-2]
- [内容][title-3] frontとapiのコードについて

[title-1]: #最初のdeploy
[title-2]: #ci
[title-3]: #内容

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
$ ecs-cli configure profile --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY --profile-name fargate-sample
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
- DNS nameはfrontからapiにアクセスするときに使うので、VUE_APP_API_HOSTに設定しておく(`http://xxx.com`の形にしておく)
```
$ export VUE_APP_API_HOST="XXX"
```
- 次に、docker imageをECRに登録してpushする。
- ここではtag: latestを使うので、docker-compose.ecs.ymlのimageで`$CIRCLE_SHA1`の行をコメントアウトして、latestのタグの行を使う。最初のデプロイだけlatestなので、service完了してCIの準備もできたら元に戻す。
- まずはfrontのコンテナから
- AWS_ACCOUNT_IDはMy Accountから見れる
```
$ export AWS_ACCOUNT_ID="XXX" AWS_DEFAULT_REGION="ap-northeast-1"
$ aws ecr get-login-password --region $AWS_DEFAULT_REGION | docker login --username AWS --password-stdin $AWS_ACCOUNT_ID.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com
$ docker build -t goping-front www/ --build-arg VUE_APP_API_HOST=$VUE_APP_API_HOST
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
- docker-compose.ecs.ymlでlatestタグを使っているかもう一度確認
- gopingRoleへのRole付与を行う
```
$ aws iam --region ap-northeast-1 create-role --role-name gopingRole --assume-role-policy-document file://task-role.json
$ aws iam --region ap-northeast-1 attach-role-policy --role-name gopingRole --policy-arn arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy
```
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
AWS_REGION
```
- AWS_ECR_ACCOUNT_URLは`$AWS_ACCOUNT_ID.dkr.ecr.$AWS_DEFAULT_REGION.amazonaws.com`を設定すればよい
- AWS_REGIONはCircleCI Orbsで使う。
- さらに、docker-compose.ecs.ymlでimageに`$CIRCLE_SHA1`が使われているか確認(最初のdeployでlatestを使っていたのを、`$CIRCLE_SHA1`に戻す)
- あとはGitHubにpushすればCircleCIが走る
- DNS nameにアクセスしてうまくいっているか確認
- 変だと思ったらCloudWatchを確認

## 内容
### front
- www/frontend/ で管理
- vue
- vue cli 3で生成しており、vue-routerを使用。JavaScriptで書いている
- frontend/views/Home.vueで`/`の内容を書いている。axiosでapi側から情報を取得
- frontend/router/index.jsでルーティングしている

### api
- api/ で管理
- golang(1.14)
- Echo v4を使用
- handler/handler.goにほぼすべての挙動を書いている
- apiのURLはすべて`/api/v1/xxx`という形で書く。現在は`/api/v1/all`でIPなどの情報が得られる
- サーバが立ち上がっているか確認用に`/hc`を使っている