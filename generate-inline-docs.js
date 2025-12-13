#!/usr/bin/env node
const Anthropic = require('@anthropic-ai/sdk');
const { glob } = require('glob');
const fs = require('fs').promises;
const path = require('path');

const client = new Anthropic.default();

// Configure source patterns based on project type
const SOURCE_PATTERNS = [
  // === MAIN REPO ===
  // TypeScript/JavaScript
  'src/**/*.ts', 'src/**/*.tsx', 'lib/**/*.ts', 'app/**/*.ts', 'app/**/*.tsx',
  // Java
  'src/**/*.java', '**/src/main/java/**/*.java',
  // Python
  'src/**/*.py', 'lib/**/*.py', '**/*.py',
  // Rust
  'src/**/*.rs',

  // === DEPENDENCY REPOS (cloned to ../deps/) ===
  // TypeScript/JavaScript
  '../deps/**/src/**/*.ts', '../deps/**/src/**/*.tsx', '../deps/**/lib/**/*.ts',
  // Java
  '../deps/**/src/main/java/**/*.java',
  // Python
  '../deps/**/src/**/*.py', '../deps/**/lib/**/*.py',
  // Rust
  '../deps/**/src/**/*.rs',

  // === EXCLUSIONS ===
  '!**/*.test.*', '!**/*.spec.*', '!**/node_modules/**',
  '!**/target/**', '!**/build/**', '!**/__pycache__/**',
  '!**/dist/**', '!**/.next/**', '!**/coverage/**'
];

// Skip generating docs for files that already have .md files
async function hasExistingDoc(filePath) {
  const docPath = filePath.replace(/\.[^.]+$/, '.md');
  try {
    await fs.access(docPath);
    return true;
  } catch {
    return false;
  }
}

async function generateInlineDoc(filePath) {
  const content = await fs.readFile(filePath, 'utf-8');
  const ext = path.extname(filePath);
  const fileName = path.basename(filePath, ext);
  const docPath = filePath.replace(/\.[^.]+$/, '.md');

  // Skip if content is too short (likely a simple export/index file)
  if (content.length < 100) {
    console.log(`  ⏭️  Skipping ${fileName} (too short)`);
    return null;
  }

  // Determine language for code blocks
  const langMap = {
    '.ts': 'typescript', '.tsx': 'typescript',
    '.js': 'javascript', '.jsx': 'javascript',
    '.java': 'java', '.py': 'python', '.rs': 'rust'
  };
  const language = langMap[ext] || 'plaintext';

  const response = await client.messages.create({
    model: 'claude-sonnet-4-20250514',
    max_tokens: 2048,
    system: `You are a technical documentation expert. Generate concise, developer-friendly documentation for source code files.

FORMAT REQUIREMENTS:
- Start with a brief 1-2 sentence description of the file's purpose
- Include a "## Key Components" section listing main exports/classes/functions
- Include a "## Usage Example" section with practical code examples
- Keep it concise - aim for 100-300 words total
- Use ${language} code blocks for examples
- Do not include the file name as a heading (it will be in the filename)`,
    messages: [{
      role: 'user',
      content: `Document this ${language} file "${fileName}${ext}":\n\n\`\`\`${language}\n${content.slice(0, 12000)}\n\`\`\``
    }]
  });

  const docContent = response.content[0].text;
  await fs.writeFile(docPath, docContent);
  console.log(`  ✅ Generated: ${docPath}`);
  return docPath;
}

async function main() {
  console.log('🔍 Scanning for source files...');

  // Find all matching files
  const files = await glob(SOURCE_PATTERNS, { ignore: ['**/node_modules/**'] });
  console.log(`Found ${files.length} source files`);

  let generated = 0;
  let skipped = 0;
  let errors = 0;

  for (const file of files) {
    try {
      // Skip if doc already exists
      if (await hasExistingDoc(file)) {
        console.log(`  ⏭️  Skipping ${file} (doc exists)`);
        skipped++;
        continue;
      }

      const result = await generateInlineDoc(file);
      if (result) generated++;
      else skipped++;

      // Rate limiting - wait 500ms between API calls
      await new Promise(r => setTimeout(r, 500));
    } catch (error) {
      console.error(`  ❌ Error processing ${file}:`, error.message);
      errors++;
    }
  }

  console.log(`\n📊 Stage 1 Complete:`);
  console.log(`   Generated: ${generated}`);
  console.log(`   Skipped: ${skipped}`);
  console.log(`   Errors: ${errors}`);

  // Write stats for later stages
  await fs.writeFile('.doc-stage1-stats.json', JSON.stringify({
    generated, skipped, errors, timestamp: new Date().toISOString()
  }));
}

main().catch(err => {
  console.error('Stage 1 failed:', err);
  process.exit(1);
});
