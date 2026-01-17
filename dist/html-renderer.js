"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.HtmlRenderer = void 0;
const playwright_1 = require("playwright");
class HtmlRenderer {
    constructor(config) {
        this.browser = null;
        this.width = config.render.width;
        this.height = config.render.height;
    }
    async initialize() {
        if (!this.browser) {
            this.browser = await playwright_1.chromium.launch({
                headless: true,
                args: ['--no-sandbox', '--disable-setuid-sandbox'],
            });
        }
    }
    async render(html, outputPath) {
        if (!this.browser) {
            await this.initialize();
        }
        const page = await this.browser.newPage();
        try {
            // Set viewport to exact dimensions
            await page.setViewportSize({
                width: this.width,
                height: this.height,
            });
            // Load HTML content
            await page.setContent(html, { waitUntil: 'networkidle' });
            // Wait a bit for any dynamic content
            await page.waitForTimeout(500);
            // Take screenshot
            await page.screenshot({
                path: outputPath,
                type: 'png',
                fullPage: false,
            });
        }
        finally {
            await page.close();
        }
    }
    async close() {
        if (this.browser) {
            await this.browser.close();
            this.browser = null;
        }
    }
}
exports.HtmlRenderer = HtmlRenderer;
