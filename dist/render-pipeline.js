"use strict";
var __createBinding = (this && this.__createBinding) || (Object.create ? (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    var desc = Object.getOwnPropertyDescriptor(m, k);
    if (!desc || ("get" in desc ? !m.__esModule : desc.writable || desc.configurable)) {
      desc = { enumerable: true, get: function() { return m[k]; } };
    }
    Object.defineProperty(o, k2, desc);
}) : (function(o, m, k, k2) {
    if (k2 === undefined) k2 = k;
    o[k2] = m[k];
}));
var __setModuleDefault = (this && this.__setModuleDefault) || (Object.create ? (function(o, v) {
    Object.defineProperty(o, "default", { enumerable: true, value: v });
}) : function(o, v) {
    o["default"] = v;
});
var __importStar = (this && this.__importStar) || (function () {
    var ownKeys = function(o) {
        ownKeys = Object.getOwnPropertyNames || function (o) {
            var ar = [];
            for (var k in o) if (Object.prototype.hasOwnProperty.call(o, k)) ar[ar.length] = k;
            return ar;
        };
        return ownKeys(o);
    };
    return function (mod) {
        if (mod && mod.__esModule) return mod;
        var result = {};
        if (mod != null) for (var k = ownKeys(mod), i = 0; i < k.length; i++) if (k[i] !== "default") __createBinding(result, mod, k[i]);
        __setModuleDefault(result, mod);
        return result;
    };
})();
Object.defineProperty(exports, "__esModule", { value: true });
exports.RenderPipeline = void 0;
const fs = __importStar(require("fs/promises"));
const path = __importStar(require("path"));
const data_collector_1 = require("./data-collector");
const template_renderer_1 = require("./template-renderer");
const html_renderer_1 = require("./html-renderer");
const image_converter_1 = require("./image-converter");
class RenderPipeline {
    constructor(config) {
        this.stats = null;
        this.config = config;
        this.dataCollector = new data_collector_1.DataCollector(config.dataSources.jsonFiles, config.dataSources.apiEndpoints, config.dataSources.scripts);
        this.templateRenderer = new template_renderer_1.TemplateRenderer(config.paths.template);
        this.htmlRenderer = new html_renderer_1.HtmlRenderer(config);
        this.imageConverter = new image_converter_1.ImageConverter(config);
    }
    async initialize() {
        // Ensure output directory exists
        const outputDir = path.dirname(this.config.render.outputPath);
        try {
            await fs.mkdir(outputDir, { recursive: true });
        }
        catch (error) {
            // Directory might already exist
        }
        await this.htmlRenderer.initialize();
    }
    async render() {
        const startTime = Date.now();
        try {
            // Step 1: Collect data
            const dataStart = Date.now();
            const viewModel = await this.dataCollector.collect();
            const dataFetchDuration = Date.now() - dataStart;
            // Step 2: Render HTML template
            const templateStart = Date.now();
            const html = await this.templateRenderer.render(viewModel);
            const renderDuration = Date.now() - templateStart;
            // Step 3: Render HTML to PNG using headless browser
            const tempPngPath = this.config.render.tempPath.replace('.tmp', '.png');
            await this.htmlRenderer.render(html, tempPngPath);
            // Step 4: Convert PNG to BMP3 1-bit
            const conversionStart = Date.now();
            await this.imageConverter.convertPngToBmp(tempPngPath);
            const conversionDuration = Date.now() - conversionStart;
            // Step 5: Atomic file swap is handled by ImageMagick or fallback
            // The converter writes to outputPath directly, or creates PNG fallback
            const finalPath = this.config.render.outputPath;
            // Get output file size
            let outputSize = 0;
            try {
                const stats = await fs.stat(finalPath);
                outputSize = stats.size;
            }
            catch {
                // File might not exist yet
            }
            // Clean up temp files
            try {
                await fs.unlink(tempPngPath);
            }
            catch {
                // Ignore cleanup errors
            }
            const totalDuration = Date.now() - startTime;
            this.stats = {
                dataFetchDuration,
                renderDuration,
                conversionDuration,
                outputPath: finalPath,
                outputSize,
                lastRenderTime: new Date(),
            };
            console.log(`Render complete:`, {
                totalDuration: `${totalDuration}ms`,
                dataFetch: `${dataFetchDuration}ms`,
                render: `${renderDuration}ms`,
                conversion: `${conversionDuration}ms`,
                outputSize: `${(outputSize / 1024).toFixed(2)} KB`,
                outputPath: finalPath,
            });
            return this.stats;
        }
        catch (error) {
            console.error('Render pipeline error:', error);
            throw error;
        }
    }
    getStats() {
        return this.stats;
    }
    async close() {
        await this.htmlRenderer.close();
    }
}
exports.RenderPipeline = RenderPipeline;
