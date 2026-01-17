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
const express_1 = __importDefault(require("express"));
const fs = __importStar(require("fs/promises"));
const path = __importStar(require("path"));
const cron = __importStar(require("node-cron"));
const render_pipeline_1 = require("./render-pipeline");
const config = require('../config.json');
const app = (0, express_1.default)();
const renderPipeline = new render_pipeline_1.RenderPipeline(config);
// CORS middleware for local LAN access
app.use((req, res, next) => {
    res.header('Access-Control-Allow-Origin', '*');
    res.header('Access-Control-Allow-Methods', 'GET, POST, OPTIONS');
    res.header('Access-Control-Allow-Headers', 'Content-Type, Access-Token');
    next();
});
// TRMNL /api/setup endpoint
app.get('/api/setup', (req, res) => {
    const baseUrl = `${req.protocol}://${req.get('host')}`;
    const imageUrl = `${baseUrl}/screen.bmp`;
    res.json({
        api_key: config.trmnl.apiKey,
        friendly_id: config.trmnl.friendlyId,
        image_url: imageUrl,
    });
});
// TRMNL /api/display endpoint
app.get('/api/display', (req, res) => {
    const token = req.headers['access-token'];
    if (token !== config.trmnl.apiKey) {
        return res.status(401).json({ error: 'Invalid Access-Token' });
    }
    const baseUrl = `${req.protocol}://${req.get('host')}`;
    const imageUrl = `${baseUrl}/screen.bmp`;
    res.json({
        image_url: imageUrl,
        refresh_rate: String(config.trmnl.refreshRateSeconds),
    });
});
// Serve screen.bmp (or screen.png as fallback)
app.get('/screen.bmp', async (req, res) => {
    try {
        const bmpPath = path.resolve(config.render.outputPath);
        const pngPath = bmpPath.replace('.bmp', '.png');
        // Try BMP first, fallback to PNG
        let filePath = bmpPath;
        let contentType = 'image/bmp';
        try {
            await fs.access(bmpPath);
        }
        catch {
            // BMP doesn't exist, try PNG
            try {
                await fs.access(pngPath);
                filePath = pngPath;
                contentType = 'image/png';
            }
            catch {
                return res.status(404).json({ error: 'Image not yet generated' });
            }
        }
        // Set headers to prevent caching
        res.setHeader('Content-Type', contentType);
        res.setHeader('Cache-Control', 'no-cache, no-store, must-revalidate');
        res.setHeader('Pragma', 'no-cache');
        res.setHeader('Expires', '0');
        // Set ETag based on file mtime
        try {
            const stats = await fs.stat(filePath);
            const etag = `"${stats.mtime.getTime()}"`;
            res.setHeader('ETag', etag);
            // Check if client has matching ETag
            if (req.headers['if-none-match'] === etag) {
                return res.status(304).end();
            }
        }
        catch {
            // Ignore ETag errors
        }
        // Send file
        res.sendFile(filePath);
    }
    catch (error) {
        console.error('Error serving image:', error);
        res.status(500).json({ error: 'Failed to serve image' });
    }
});
// Serve screen.png (alias for compatibility)
app.get('/screen.png', async (req, res) => {
    // Reuse the same logic as screen.bmp
    try {
        const bmpPath = path.resolve(config.render.outputPath);
        const pngPath = bmpPath.replace('.bmp', '.png');
        // Try PNG first, then BMP
        let filePath = pngPath;
        let contentType = 'image/png';
        try {
            await fs.access(pngPath);
        }
        catch {
            // PNG doesn't exist, try BMP
            try {
                await fs.access(bmpPath);
                filePath = bmpPath;
                contentType = 'image/bmp';
            }
            catch {
                return res.status(404).json({ error: 'Image not yet generated' });
            }
        }
        res.setHeader('Content-Type', contentType);
        res.setHeader('Cache-Control', 'no-cache, no-store, must-revalidate');
        res.setHeader('Pragma', 'no-cache');
        res.setHeader('Expires', '0');
        try {
            const stats = await fs.stat(filePath);
            const etag = `"${stats.mtime.getTime()}"`;
            res.setHeader('ETag', etag);
            if (req.headers['if-none-match'] === etag) {
                return res.status(304).end();
            }
        }
        catch {
            // Ignore ETag errors
        }
        res.sendFile(filePath);
    }
    catch (error) {
        console.error('Error serving image:', error);
        res.status(500).json({ error: 'Failed to serve image' });
    }
});
// Manual render trigger endpoint
app.post('/api/render', async (req, res) => {
    try {
        const stats = await renderPipeline.render();
        res.json({
            success: true,
            stats,
        });
    }
    catch (error) {
        console.error('Manual render failed:', error);
        res.status(500).json({
            success: false,
            error: error.message,
        });
    }
});
// Health/status endpoint
app.get('/api/status', (req, res) => {
    const stats = renderPipeline.getStats();
    res.json({
        status: 'running',
        config: {
            refreshIntervalMinutes: config.render.refreshIntervalMinutes,
            outputPath: config.render.outputPath,
        },
        lastRender: stats ? {
            time: stats.lastRenderTime.toISOString(),
            dataFetchDuration: stats.dataFetchDuration,
            renderDuration: stats.renderDuration,
            conversionDuration: stats.conversionDuration,
            outputSize: stats.outputSize,
        } : null,
    });
});
// Root endpoint
app.get('/', (req, res) => {
    const baseUrl = `${req.protocol}://${req.get('host')}`;
    res.json({
        service: 'TRMNL Renderer',
        endpoints: {
            setup: `${baseUrl}/api/setup`,
            display: `${baseUrl}/api/display`,
            image: `${baseUrl}/screen.bmp`,
            render: `${baseUrl}/api/render (POST)`,
            status: `${baseUrl}/api/status`,
        },
    });
});
async function main() {
    const port = config.server.port;
    const host = config.server.host;
    // Initialize render pipeline
    await renderPipeline.initialize();
    // Perform initial render
    console.log('Performing initial render...');
    try {
        await renderPipeline.render();
    }
    catch (error) {
        console.error('Initial render failed:', error);
    }
    // Schedule periodic renders
    const intervalMinutes = config.render.refreshIntervalMinutes;
    const cronExpression = `*/${intervalMinutes} * * * *`;
    console.log(`Scheduling renders every ${intervalMinutes} minutes`);
    cron.schedule(cronExpression, async () => {
        console.log('Scheduled render triggered');
        try {
            await renderPipeline.render();
        }
        catch (error) {
            console.error('Scheduled render failed:', error);
        }
    });
    // Start server
    app.listen(port, host, () => {
        console.log(`TRMNL Renderer service listening on http://${host === '0.0.0.0' ? 'localhost' : host}:${port}`);
        console.log(`Endpoints:`);
        console.log(`  - TRMNL Setup: http://${host === '0.0.0.0' ? 'localhost' : host}:${port}/api/setup`);
        console.log(`  - TRMNL Display: http://${host === '0.0.0.0' ? 'localhost' : host}:${port}/api/display`);
        console.log(`  - Image: http://${host === '0.0.0.0' ? 'localhost' : host}:${port}/screen.bmp`);
        console.log(`  - Manual Render: POST http://${host === '0.0.0.0' ? 'localhost' : host}:${port}/api/render`);
    });
    // Graceful shutdown
    process.on('SIGTERM', async () => {
        console.log('Shutting down...');
        await renderPipeline.close();
        process.exit(0);
    });
    process.on('SIGINT', async () => {
        console.log('Shutting down...');
        await renderPipeline.close();
        process.exit(0);
    });
}
main().catch(error => {
    console.error('Fatal error:', error);
    process.exit(1);
});
