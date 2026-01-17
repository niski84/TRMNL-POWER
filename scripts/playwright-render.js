#!/usr/bin/env node

/**
 * Playwright HTML to PNG renderer
 * Converts HTML content to PNG image using Playwright
 * 
 * Usage:
 *   node playwright-render.js <html-file> <output-png> <width> <height>
 *   OR
 *   cat html.txt | node playwright-render.js - <output-png> <width> <height>
 */

const playwright = require('playwright');
const fs = require('fs');
const path = require('path');

async function renderHTML(html, outputPath, width, height) {
    let browser = null;
    try {
        // Launch browser
        browser = await playwright.chromium.launch({
            headless: true,
            args: [
                '--no-sandbox',
                '--disable-setuid-sandbox',
                '--disable-dev-shm-usage',
                '--disable-gpu',
            ],
        });

        const page = await browser.newPage();

        // Set viewport to exact dimensions
        await page.setViewportSize({
            width: parseInt(width, 10),
            height: parseInt(height, 10),
        });

        // Load HTML content
        await page.setContent(html, {
            waitUntil: 'networkidle',
            timeout: 30000, // 30 second timeout
        });

        // Wait a bit for any dynamic content or animations
        await page.waitForTimeout(500);

        // Take screenshot
        await page.screenshot({
            path: outputPath,
            type: 'png',
            fullPage: false,
        });

        await page.close();
        return true;
    } catch (error) {
        console.error('Playwright render error:', error.message);
        process.stderr.write(`Error: ${error.message}\n`);
        return false;
    } finally {
        if (browser) {
            await browser.close();
        }
    }
}

// Main execution
async function main() {
    const args = process.argv.slice(2);

    if (args.length < 4) {
        console.error('Usage: node playwright-render.js <html-file-or-> <output-png> <width> <height>');
        console.error('  Use "-" as html-file to read from stdin');
        process.exit(1);
    }

    const [htmlSource, outputPath, width, height] = args;

    // Read HTML content
    let html;
    if (htmlSource === '-') {
        // Read from stdin
        html = fs.readFileSync(0, 'utf-8');
    } else {
        // Read from file
        if (!fs.existsSync(htmlSource)) {
            console.error(`Error: HTML file not found: ${htmlSource}`);
            process.exit(1);
        }
        html = fs.readFileSync(htmlSource, 'utf-8');
    }

    // Ensure output directory exists
    const outputDir = path.dirname(outputPath);
    if (!fs.existsSync(outputDir)) {
        fs.mkdirSync(outputDir, { recursive: true });
    }

    // Render
    const success = await renderHTML(html, outputPath, width, height);

    if (!success) {
        process.exit(1);
    }
}

// Run if executed directly
if (require.main === module) {
    main().catch(error => {
        console.error('Fatal error:', error);
        process.exit(1);
    });
}

module.exports = { renderHTML };

