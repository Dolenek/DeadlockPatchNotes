export function cleanSteamContent(contents) {
  return contents
    .replace(/\\r/g, "")
    .replace(/\[p\]\[\/p\]/g, "\n")
    .replace(/\[p\]/g, "")
    .replace(/\[\/p\]/g, "\n")
    .replace(/\[b\]|\[\/b\]|\[u\]|\[\/u\]/g, "")
    .replace(/\\\[/g, "[")
    .replace(/\\\]/g, "]");
}

export function splitSections(lines) {
  const sections = [];
  let current = null;

  for (const line of lines) {
    const match = line.match(/^[\[\]]\s+(.+?)\s+\]$/);
    if (match) {
      current = { name: match[1], lines: [] };
      sections.push(current);
      continue;
    }

    if (current) {
      current.lines.push(line);
    }
  }

  return sections;
}

export function parseBullet(line) {
  const match = line.match(/^-\s*([^:]+):\s*(.*)$/);
  if (!match) {
    return null;
  }
  return {
    prefix: match[1].trim(),
    text: match[2].trim(),
  };
}

export function extractNewItemName(text) {
  const cleaned = text.replace(/[.]+$/, "").trim();
  return cleaned || null;
}

export function getCoreSections(steamItem) {
  const cleaned = cleanSteamContent(steamItem.contents);
  const lines = cleaned
    .split(/\n+/)
    .map((line) => line.trim())
    .filter(Boolean);

  const sections = splitSections(lines);
  const sectionByName = new Map(sections.map((section) => [section.name.toLowerCase(), section]));

  const generalSection = sectionByName.get("general");
  const itemsSection = sectionByName.get("items");
  const heroesSection = sectionByName.get("heroes");

  if (!generalSection || !itemsSection || !heroesSection) {
    throw new Error("Expected General, Items, and Heroes sections in source patch");
  }

  return { generalSection, itemsSection, heroesSection };
}
