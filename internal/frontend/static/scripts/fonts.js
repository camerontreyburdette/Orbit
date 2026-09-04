import { invokeBackendMethod } from "./backend.js";

function buildFontFaceRule(fontDescriptor) {
  return (
    `@font-face{font-family:'OrbitFont';src:url(${fontDescriptor.data_uri}) ` +
    `format('${fontDescriptor.format}');font-weight:${fontDescriptor.weight};font-style:${fontDescriptor.style};` +
    "font-display:swap;}"
  );
}

export async function loadCustomFonts() {
  try {
    const fontsList = await invokeBackendMethod("get_fonts");
    if (!fontsList || !fontsList.length) {
      return;
    }
    const styleElement = document.createElement("style");
    styleElement.textContent =
      fontsList.map(buildFontFaceRule).join("\n") + "\nbody{font-family:'OrbitFont','Segoe UI',system-ui,sans-serif;}";
    document.head.append(styleElement);
  } catch {}
}
