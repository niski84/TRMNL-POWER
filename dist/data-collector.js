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
exports.DataCollector = void 0;
const fs = __importStar(require("fs/promises"));
class DataCollector {
    constructor(jsonFiles, apiEndpoints, scripts) {
        this.jsonFiles = jsonFiles;
        this.apiEndpoints = apiEndpoints;
        this.scripts = scripts;
    }
    async collect() {
        const rawData = {};
        // Collect from JSON files
        for (const filePath of this.jsonFiles) {
            try {
                const content = await fs.readFile(filePath, 'utf-8');
                const data = JSON.parse(content);
                Object.assign(rawData, data);
            }
            catch (error) {
                console.warn(`Failed to read JSON file ${filePath}:`, error);
            }
        }
        // Collect from API endpoints
        for (const endpoint of this.apiEndpoints) {
            try {
                const response = await fetch(endpoint);
                if (response.ok) {
                    const data = await response.json();
                    Object.assign(rawData, data);
                }
            }
            catch (error) {
                console.warn(`Failed to fetch from API ${endpoint}:`, error);
            }
        }
        // Collect from scripts (would execute and parse JSON output)
        // This is a placeholder - actual implementation would exec scripts
        for (const scriptPath of this.scripts) {
            try {
                // In production, you'd execute the script and capture JSON output
                console.warn(`Script execution not yet implemented for ${scriptPath}`);
            }
            catch (error) {
                console.warn(`Failed to execute script ${scriptPath}:`, error);
            }
        }
        // Normalize into ViewModel
        return this.normalize(rawData);
    }
    normalize(rawData) {
        // Default values with placeholders
        const title = rawData.title || 'TRMNL Dashboard';
        const timestamp = rawData.timestamp || new Date().toLocaleString();
        const cards = [];
        // Extract card data with safe defaults
        // Support common patterns: cards array, or individual card fields
        if (Array.isArray(rawData.cards)) {
            for (const card of rawData.cards) {
                cards.push({
                    label: card.label || 'N/A',
                    value: card.value ?? 'N/A',
                    unit: card.unit || '',
                    trend: card.trend || 'neutral',
                });
            }
        }
        // If no cards array, try to extract individual fields as cards
        if (cards.length === 0) {
            const cardKeys = Object.keys(rawData).filter(key => !['title', 'timestamp', 'cards'].includes(key));
            for (let i = 0; i < Math.min(4, cardKeys.length); i++) {
                const key = cardKeys[i];
                const value = rawData[key];
                cards.push({
                    label: this.formatLabel(key),
                    value: this.formatValue(value),
                    unit: '',
                    trend: 'neutral',
                });
            }
        }
        // Ensure at least 2 cards for layout
        while (cards.length < 2) {
            cards.push({
                label: 'Placeholder',
                value: 'N/A',
                unit: '',
                trend: 'neutral',
            });
        }
        // Limit to 4 cards max
        const finalCards = cards.slice(0, 4);
        return {
            title,
            timestamp,
            cards: finalCards,
        };
    }
    formatLabel(key) {
        // Convert camelCase or snake_case to Title Case
        return key
            .replace(/([A-Z])/g, ' $1')
            .replace(/_/g, ' ')
            .trim()
            .split(' ')
            .map(word => word.charAt(0).toUpperCase() + word.slice(1).toLowerCase())
            .join(' ');
    }
    formatValue(value) {
        if (value === null || value === undefined) {
            return 'N/A';
        }
        if (typeof value === 'number') {
            return value;
        }
        if (typeof value === 'boolean') {
            return value ? 'Yes' : 'No';
        }
        const str = String(value);
        // Truncate very long strings
        if (str.length > 20) {
            return str.substring(0, 17) + '...';
        }
        return str;
    }
}
exports.DataCollector = DataCollector;
