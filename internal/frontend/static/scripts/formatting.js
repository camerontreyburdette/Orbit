const KILOBYTE = 1024;
const MEGABYTE = 1048576;
const GIGABYTE = 1073741824;

export function formatByteSize(byteCount) {
  if (byteCount < KILOBYTE) {
    return byteCount + " B";
  }
  if (byteCount < MEGABYTE) {
    return (byteCount / KILOBYTE).toFixed(1) + " KB";
  }
  if (byteCount < GIGABYTE) {
    return (byteCount / MEGABYTE).toFixed(1) + " MB";
  }
  return (byteCount / GIGABYTE).toFixed(2) + " GB";
}

export function formatDateString(timestampString) {
  if (!timestampString) {
    return "";
  }
  const parsedDate = new Date(String(timestampString).replace(" ", "T"));
  if (isNaN(parsedDate)) {
    return String(timestampString).split(" ")[0];
  }
  return parsedDate.getMonth() + 1 + "/" + parsedDate.getDate() + "/" + parsedDate.getFullYear();
}

const SECONDS_PER_MINUTE = 60;
const SECONDS_PER_HOUR = 3600;

export function formatTimeSpent(totalSeconds) {
  const safeSeconds = Math.max(0, Number(totalSeconds) || 0);
  const hours = Math.floor(safeSeconds / SECONDS_PER_HOUR);
  const minutes = Math.floor((safeSeconds % SECONDS_PER_HOUR) / SECONDS_PER_MINUTE);
  if (hours === 0) {
    return minutes + "m";
  }
  return hours + "h " + minutes + "m";
}

export function extractDatePortion(timestampString) {
  return String(timestampString || "").split(" ")[0];
}
