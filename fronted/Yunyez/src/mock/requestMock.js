
import axios from 'axios'

const service = axios.create({
  baseURL: 'http://127.0.0.1:4523/m1/3339466-2916181-default',
  timeout: 10000
})

export default service