#!/usr/bin/env node
const Anthropic = require('@anthropic-ai/sdk');
const fs = require('fs').promises;
const path = require('path');
const { glob } = require('glob');

const client = new Anthropic.default();

async function gatherContext() {
  const context = { readme: '', claudeMd: '', inlineDocs: [], codewiki: [] };

  // Read README
  try {
    context.readme = await fs.readFile('README.md', 'utf-8');
  } catch {}

  // Read CLAUDE.md
  try {
    context.claudeMd = await fs.readFile('CLAUDE.md', 'utf-8');
  } catch {}

  // Gather inline docs from Stage 1
  const inlineMdFiles = await glob('src/**/*.md');
  for (const file of inlineMdFiles.slice(0, 20)) {
    try {
      const content = await fs.readFile(file, 'utf-8');
      context.inlineDocs.push({ path: file, content: content.slice(0, 1000) });
    } catch {}
  }

  // Gather CodeWiki docs from Stage 2
  const codewikiFiles = await glob('docs/codewiki/**/*.md');
  for (const file of codewikiFiles.slice(0, 10)) {
    try {
      const content = await fs.readFile(file, 'utf-8');
      context.codewiki.push({ path: file, content: content.slice(0, 2000) });
    } catch {}
  }

  return context;
}

async function generateTutorial(title, prompt, context, outputPath) {
  console.log(`  📝 Generating: ${title}...`);

  const contextSummary = `
PROJECT CONTEXT:
${context.readme.slice(0, 3000)}

${context.claudeMd ? `PROJECT INSTRUCTIONS (CLAUDE.md):\n${context.claudeMd.slice(0, 2000)}` : ''}

INLINE DOCUMENTATION (${context.inlineDocs.length} files):
${context.inlineDocs.map(d => `- ${d.path}`).join('\n')}

CODEWIKI DOCUMENTATION (${context.codewiki.length} files):
${context.codewiki.map(d => `- ${d.path}`).join('\n')}
`;

  const response = await client.messages.create({
    model: 'claude-sonnet-4-20250514',
    max_tokens: 4096,
    system: `You are a technical writer creating developer documentation.
Write clear, practical documentation that developers can follow step-by-step.
Use Markdown formatting with proper headings, code blocks, and examples.
Be concise but thorough. Include real code examples when helpful.`,
    messages: [{
      role: 'user',
      content: `${prompt}\n\n${contextSummary}`
    }]
  });

  const content = response.content[0].text;
  await fs.mkdir(path.dirname(outputPath), { recursive: true });
  await fs.writeFile(outputPath, content);
  console.log(`  ✅ Written: ${outputPath}`);
  return outputPath;
}

async function main() {
  console.log('🔍 Gathering project context...');
  const context = await gatherContext();

  console.log(`  README: ${context.readme.length} chars`);
  console.log(`  CLAUDE.md: ${context.claudeMd.length} chars`);
  console.log(`  Inline docs: ${context.inlineDocs.length} files`);
  console.log(`  CodeWiki docs: ${context.codewiki.length} files`);

  const tutorials = [
    {
      title: 'Getting Started',
      outputPath: 'docs/tutorials/getting-started.md',
      prompt: `Write a "Getting Started" guide for this project.
Include:
1. Prerequisites (what needs to be installed)
2. Installation steps
3. Basic configuration
4. Running the project locally
5. First steps / "Hello World" equivalent
6. Common issues and solutions

Make it beginner-friendly but assume the reader knows basic programming.`
    },
    {
      title: 'Architecture Overview',
      outputPath: 'docs/tutorials/architecture-overview.md',
      prompt: `Write an "Architecture Overview" document for this project.
Include:
1. High-level architecture diagram description (describe what a diagram would show)
2. Main components and their responsibilities
3. Data flow between components
4. Key design decisions and patterns used
5. Directory/folder structure explanation

Focus on helping new developers understand the codebase structure.`
    },
    {
      title: 'API Examples',
      outputPath: 'docs/tutorials/api-examples.md',
      prompt: `Write an "API Usage Examples" guide for this project.
Include:
1. Main API endpoints or functions available
2. Authentication (if applicable)
3. Common use cases with code examples
4. Error handling patterns
5. Best practices

Include practical, copy-paste ready code examples.`
    }
  ];

  let generated = 0;
  for (const tutorial of tutorials) {
    try {
      await generateTutorial(
        tutorial.title,
        tutorial.prompt,
        context,
        tutorial.outputPath
      );
      generated++;
      // Rate limiting
      await new Promise(r => setTimeout(r, 1000));
    } catch (error) {
      console.error(`  ❌ Failed: ${tutorial.title}:`, error.message);
    }
  }

  console.log(`\n📊 Stage 3 Complete:`);
  console.log(`   Generated: ${generated} tutorials`);

  await fs.writeFile('.doc-stage3-stats.json', JSON.stringify({
    generated, timestamp: new Date().toISOString()
  }));
}

main().catch(err => {
  console.error('Stage 3 failed:', err);
  process.exit(1);
});
