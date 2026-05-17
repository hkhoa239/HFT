declare global {
  interface Window {
    APP_CONFIG: {
      API_URL: string;
      APP_ENV: string;
      ENABLE_DEBUG: string;
      VERSION: string;
    };
  }
}

const runtimeConfig = window.APP_CONFIG || {};

export const APP_CONFIG = {
  apiUrl: runtimeConfig.API_URL || 'http://127.0.0.1:8080',
  env: runtimeConfig.APP_ENV || 'local',
  debug: runtimeConfig.ENABLE_DEBUG === 'true',
  version: runtimeConfig.VERSION || '1.0.0-mvp.baseline'
};
