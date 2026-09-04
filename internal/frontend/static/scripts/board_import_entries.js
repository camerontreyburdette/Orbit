const BOARD_DOCUMENT_FILE_NAME = "board.json";
const ATTACHMENTS_DIRECTORY_NAME = "attachments";
const JSON_EXTENSION = ".json";

function readDirectoryBatch(directoryReader) {
  return new Promise((resolve, reject) => directoryReader.readEntries(resolve, reject));
}

async function readAllDirectoryEntries(directoryEntry) {
  const directoryReader = directoryEntry.createReader();
  const allEntries = [];
  let batch = await readDirectoryBatch(directoryReader);
  while (batch.length) {
    allEntries.push(...batch);
    batch = await readDirectoryBatch(directoryReader);
  }
  return allEntries;
}

function readFileEntry(fileEntry) {
  return new Promise((resolve, reject) => fileEntry.file(resolve, reject));
}

async function collectFilesFromEntry(entry, relativeDirectory, collectedFiles) {
  if (entry.isFile) {
    collectedFiles.push({ relativePath: relativeDirectory + entry.name, file: await readFileEntry(entry) });
    return;
  }
  if (!entry.isDirectory) {
    return;
  }
  const childDirectory = relativeDirectory + entry.name + "/";
  for (const childEntry of await readAllDirectoryEntries(entry)) {
    await collectFilesFromEntry(childEntry, childDirectory, collectedFiles);
  }
}

function extractFileSystemEntries(dataTransfer) {
  return [...(dataTransfer.items || [])]
    .map((item) => (typeof item.webkitGetAsEntry === "function" ? item.webkitGetAsEntry() : null))
    .filter(Boolean);
}

export function collectPickedFiles(fileList) {
  return [...fileList].map((file) => ({ relativePath: file.webkitRelativePath || file.name, file }));
}

export async function collectDroppedFiles(dataTransfer) {
  const entries = extractFileSystemEntries(dataTransfer);
  if (!entries.length) {
    return collectPickedFiles(dataTransfer.files);
  }
  const collectedFiles = [];
  for (const entry of entries) {
    await collectFilesFromEntry(entry, "", collectedFiles);
  }
  return collectedFiles;
}

function parentDirectoryOf(relativePath) {
  const slashIndex = relativePath.lastIndexOf("/");
  return slashIndex < 0 ? "" : relativePath.slice(0, slashIndex + 1);
}

function fileNameOf(relativePath) {
  return relativePath.slice(relativePath.lastIndexOf("/") + 1);
}

function lastSegmentOf(directoryPath) {
  return fileNameOf(directoryPath.slice(0, -1));
}

function isBoardDocument(collectedFile) {
  return fileNameOf(collectedFile.relativePath).toLowerCase() === BOARD_DOCUMENT_FILE_NAME;
}

function isLooseJsonDocument(collectedFile) {
  return !isBoardDocument(collectedFile) && collectedFile.relativePath.toLowerCase().endsWith(JSON_EXTENSION);
}

function collectAttachmentsUnder(collectedFiles, rootDirectory) {
  const attachmentPrefix = rootDirectory + ATTACHMENTS_DIRECTORY_NAME + "/";
  return collectedFiles
    .filter((collectedFile) => collectedFile.relativePath.startsWith(attachmentPrefix))
    .map((collectedFile) => ({ relativePath: collectedFile.relativePath.slice(attachmentPrefix.length), file: collectedFile.file }));
}

function createBoardPackage(documentFile, label, attachments) {
  return { documentFile, label, attachments };
}

function createFolderBoardPackage(collectedFiles, documentEntry) {
  const rootDirectory = parentDirectoryOf(documentEntry.relativePath);
  const label = rootDirectory ? lastSegmentOf(rootDirectory) : documentEntry.file.name;
  return createBoardPackage(documentEntry.file, label, collectAttachmentsUnder(collectedFiles, rootDirectory));
}

export function groupFilesIntoBoardPackages(collectedFiles) {
  const documentEntries = collectedFiles.filter(isBoardDocument);
  const boardDirectories = new Set(documentEntries.map((documentEntry) => parentDirectoryOf(documentEntry.relativePath)));
  const folderPackages = documentEntries.map((documentEntry) => createFolderBoardPackage(collectedFiles, documentEntry));
  const loosePackages = collectedFiles
    .filter((collectedFile) => isLooseJsonDocument(collectedFile) && !boardDirectories.has(parentDirectoryOf(collectedFile.relativePath)))
    .map((collectedFile) => createBoardPackage(collectedFile.file, collectedFile.file.name, []));
  return [...folderPackages, ...loosePackages];
}
