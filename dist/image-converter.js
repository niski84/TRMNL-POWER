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
var __importDefault = (this && this.__importDefault) || function (mod) {
    return (mod && mod.__esModule) ? mod : { "default": mod };
};
Object.defineProperty(exports, "__esModule", { value: true });
exports.ImageConverter = void 0;
const sharp_1 = __importDefault(require("sharp"));
const fs = __importStar(require("fs/promises"));
class ImageConverter {
    constructor(config) {
        this.width = config.render.width;
        this.height = config.render.height;
        this.outputPath = config.render.outputPath;
    }
    async convertPngToBmp(pngPath) {
        // Convert PNG to 1-bit BMP using sharp for intermediate processing
        // Then use ImageMagick for true BMP3 1-bit if available
        try {
            // First, resize and convert to grayscale, then threshold
            const pngBuffer = await (0, sharp_1.default)(pngPath)
                .resize(this.width, this.height, {
                fit: 'fill',
                withoutEnlargement: false,
            })
                .grayscale()
                .threshold(128)
                .png({ compressionLevel: 9, colors: 2 })
                .toBuffer();
            // Save thresholded PNG as fallback
            const pngOutput = this.outputPath.replace('.bmp', '.png');
            await fs.writeFile(pngOutput, pngBuffer);
            // Try to convert to BMP using ImageMagick if available
            await this.convertToBmpIfAvailable(pngOutput);
        }
        catch (error) {
            console.error('Image conversion error:', error);
            throw error;
        }
    }
    async convertToBmpIfAvailable(pngPath) {
        // Try to use ImageMagick's convert command for true BMP3 1-bit output
        const { exec } = require('child_process');
        const { promisify } = require('util');
        const execAsync = promisify(exec);
        try {
            const bmpFile = this.outputPath;
            // Convert PNG to 1-bit BMP using ImageMagick
            // -threshold 50% converts to pure black/white
            // -depth 1 sets 1-bit depth
            // -type Bilevel ensures monochrome
            // BMP3: format specifies BMP version 3 (1-bit)
            const command = `convert "${pngPath}" -threshold 50% -type Bilevel -depth 1 -compress none BMP3:"${bmpFile}"`;
            await execAsync(command, { timeout: 10000 });
            console.log(`Successfully converted to BMP3: ${bmpFile}`);
        }
        catch (error) {
            // ImageMagick might not be available - that's okay, PNG is acceptable
            console.warn('ImageMagick conversion failed (PNG fallback will be used):', error.message);
            // Copy the thresholded PNG as .bmp for compatibility
            // Some systems can handle PNG with .bmp extension
            try {
                await fs.copyFile(pngPath, this.outputPath);
            }
            catch (copyError) {
                // Ignore copy errors
            }
        }
    }
}
exports.ImageConverter = ImageConverter;
