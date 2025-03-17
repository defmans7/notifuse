interface Config {
  API_ENDPOINT: string
}

const config: Config = {
  API_ENDPOINT: import.meta.env.VITE_API_ENDPOINT || 'http://localhost:3000'
}

export default config
