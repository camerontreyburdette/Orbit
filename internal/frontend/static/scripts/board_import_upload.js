const BOARD_FORM_FIELD_NAME = "board";
const ATTACHMENT_PATH_FORM_FIELD_NAME = "attachment_path";
const ATTACHMENT_FILE_FORM_FIELD_NAME = "attachment_file";
const BOARD_DOCUMENT_UPLOAD_NAME = "board.json";

function boardImportUniformResourceLocator() {
  return window.location.origin + "/boards/import";
}

function buildBoardImportFormData(boardPackage) {
  const formData = new FormData();
  formData.append(BOARD_FORM_FIELD_NAME, boardPackage.documentFile, BOARD_DOCUMENT_UPLOAD_NAME);
  for (const attachment of boardPackage.attachments) {
    formData.append(ATTACHMENT_PATH_FORM_FIELD_NAME, attachment.relativePath);
    formData.append(ATTACHMENT_FILE_FORM_FIELD_NAME, attachment.file, attachment.file.name);
  }
  return formData;
}

export async function uploadBoardPackage(boardPackage) {
  const response = await fetch(boardImportUniformResourceLocator(), {
    method: "POST",
    body: buildBoardImportFormData(boardPackage),
  });
  if (!response.ok) {
    throw new Error((await response.text()).trim() || "Import failed");
  }
  return response.json();
}
