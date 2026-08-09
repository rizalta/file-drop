export interface UploadDropOptions {
  formData: FormData;
  fileSize?: number;
  onProgress?: (progress: number, speed: string) => void;
  onXhrCreated?: (xhr: XMLHttpRequest) => void;
}

export const fetchServerConfig = async (): Promise<{ max_upload_size: number }> => {
  const res = await fetch("/api/config");
  if (!res.ok) {
    throw new Error("Failed to fetch server configuration");
  }
  return res.json();
};

export const uploadDrop = async ({
  formData,
  fileSize,
  onProgress,
  onXhrCreated,
}: UploadDropOptions): Promise<{ id: string }> => {
  const configData = await fetchServerConfig();
  const maxUploadSize = configData.max_upload_size;

  if (fileSize && fileSize > maxUploadSize) {
    const sizeMB = (fileSize / (1024 * 1024)).toFixed(1);
    const maxMB = (maxUploadSize / (1024 * 1024)).toFixed(0);
    throw new Error(`File size (${sizeMB} MB) exceeds maximum allowed limit of ${maxMB} MB`);
  }

  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    onXhrCreated?.(xhr);

    let lastTime = 0;
    let lastLoaded = 0;

    xhr.upload.onprogress = (event) => {
      if (event.lengthComputable && event.total > 0) {
        const pct = Math.round((event.loaded / event.total) * 100);
        let speedStr = "";

        const now = performance.now();
        if (lastTime > 0) {
          const timeDiff = (now - lastTime) / 1000;
          const bytesDiff = event.loaded - lastLoaded;

          if (timeDiff >= 0.2) {
            const bps = bytesDiff / timeDiff;
            if (bps >= 1024 * 1024) {
              speedStr = `${(bps / (1024 * 1024)).toFixed(1)} MB/s`;
            } else if (bps >= 1024) {
              speedStr = `${(bps / 1024).toFixed(0)} KB/s`;
            } else {
              speedStr = `${bps.toFixed(0)} B/s`;
            }
            lastTime = now;
            lastLoaded = event.loaded;
          }
        } else {
          lastTime = now;
          lastLoaded = event.loaded;
        }

        onProgress?.(pct, speedStr);
      }
    };

    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        try {
          const resData = JSON.parse(xhr.responseText);
          resolve(resData);
        } catch {
          reject(new Error("Invalid response from server."));
        }
      } else {
        try {
          const errData = JSON.parse(xhr.responseText);
          reject(new Error(errData.error || `Upload failed (${xhr.status})`));
        } catch {
          reject(new Error(`Upload failed (${xhr.status})`));
        }
      }
    };

    xhr.onabort = () => {
      reject(new Error("UPLOAD_CANCELLED"));
    };

    xhr.onerror = () => {
      if (xhr.responseText) {
        try {
          const errData = JSON.parse(xhr.responseText);
          if (errData.error) {
            reject(new Error(errData.error));
            return;
          }
        } catch {
          // fallback
        }
      }
      if (xhr.status >= 400) {
        reject(new Error(`Upload failed (${xhr.status})`));
        return;
      }
      reject(new Error("Upload failed. Connection reset or file size exceeds limit."));
    };

    xhr.open("POST", "/api/upload");
    xhr.send(formData);
  });
};
