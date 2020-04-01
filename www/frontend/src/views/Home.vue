<template>
  <div class="home">
    <h1>IP Address Check Website</h1>
    <div class="list-item">
      <div class="col-1 border">IP Address</div>
      <div class="col-2 border">{{ info.ip }}</div>
    </div>
    <div class="list-item">
      <div class="col-1 border">Host</div>
      <div class="col-2 border">{{ info.host }}</div>
    </div>
    <div class="list-item">
      <div class="col-1 border">Country Code</div>
      <div class="col-2 border">{{ info.country_code }}</div>
    </div>
    <div class="list-item">
      <div class="col-1 border">UA</div>
      <div class="col-2 border">{{ info.ua }}</div>
    </div>
    <div class="list-item">
      <div class="col-1 border">Port</div>
      <div class="col-2 border">{{ info.port }}</div>
    </div>
    <div class="list-item">
      <div class="col-1 border">Language</div>
      <div class="col-2 border">{{ info.lang }}</div>
    </div>
    <div class="list-item">
      <div class="col-1 border">Encoding</div>
      <div class="col-2 border">{{ info.encoding }}</div>
    </div>
    <div class="list-item">
      <div class="col-1 border">Forwarded</div>
      <div class="col-2 border">{{ info.forwarded }}</div>
    </div>
  </div>
</template>

<script>
// @ is an alias to /src
import axios from 'axios'

export default {
  name: 'Home',
  data() {
    return {
      info: {
        ip: 'network connection failed',
        host: 'network connection failed',
        country_code: 'network connection failed',
        ua: 'network connection failed',
        port: 'network connection failed',
        lang: 'network connection failed',
        encoding: 'network connection failed',
        forwarded: 'network connection failed'
      }
    }
  },
  mounted: function() {
    axios.get('http://localhost:8001/api/v1/all')
      .then(res => {
        (this.info = res.data)
        // console.log(res)
      })
  }
}
</script>

<style scoped>
.home {
  display: flex;
  flex-direction: column;
}
.list-item {
  display: flex;
  width: 80%;
  margin: 5px auto;
}
.col-1 {
  width: 50%;
}
.col-2 {
  width: 50%;
}
</style>