const config = {
  baseUrl: import.meta.env.VITE_BASE_URL,
};

export const getWebSocketBaseUrl = (baseUrl?: string): string => {
  const configuredUrl = baseUrl || window.location.origin;
  const url = new URL(configuredUrl);

  if (url.protocol === "https:") {
    url.protocol = "wss:";
  } else if (url.protocol === "http:") {
    url.protocol = "ws:";
  } else {
    throw new Error(`Unsupported API URL protocol: ${url.protocol}`);
  }

  return url.toString().replace(/\/$/, "");
};

export default config;
