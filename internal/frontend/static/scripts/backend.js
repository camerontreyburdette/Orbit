export function isBackendAvailable() {
  return Boolean(window.pywebview && window.pywebview.api);
}

export function invokeBackendMethod(methodName, ...methodArguments) {
  if (!isBackendAvailable()) {
    return Promise.reject(new Error("Backend not ready"));
  }
  return window.pywebview.api[methodName](...methodArguments).then(
    (response) => {
      if (response && typeof response === "object" && response.error) {
        console.error(response.error);
        throw new Error(response.error);
      }
      return response;
    },
    (executionError) => {
      console.error(executionError);
      throw executionError;
    }
  );
}
