import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import './index.css'

console.log(`
# AlbinoDrought/creamy-prediction-market
Repo: https://github.com/AlbinoDrought/creamy-prediction-market
Source: ${window.location.origin}/source.tar.gz
`);

const app = createApp(App)

app.use(createPinia())
app.use(router)

app.mount('#app')
