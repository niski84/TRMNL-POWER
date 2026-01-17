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
exports.TemplateRenderer = void 0;
const fs = __importStar(require("fs/promises"));
class TemplateRenderer {
    constructor(templatePath) {
        this.templatePath = templatePath;
    }
    async render(viewModel) {
        const template = await fs.readFile(this.templatePath, 'utf-8');
        // Replace title
        let html = template.replace('{{TITLE}}', this.escapeHtml(viewModel.title));
        // Replace timestamp
        html = html.replace('{{TIMESTAMP}}', this.escapeHtml(viewModel.timestamp));
        // Render cards
        const cardsHtml = this.renderCards(viewModel.cards);
        html = html.replace('{{CARDS}}', cardsHtml);
        // Adjust grid layout class based on card count
        const cardCount = viewModel.cards.length;
        let contentClass = 'content';
        if (cardCount === 1) {
            contentClass += ' single';
        }
        else if (cardCount === 3) {
            contentClass += ' three';
        }
        html = html.replace(/<div class="content"[^>]*>/, `<div class="${contentClass}" id="content">`);
        return html;
    }
    renderCards(cards) {
        return cards
            .map(card => {
            const valueDisplay = typeof card.value === 'number'
                ? card.value.toLocaleString()
                : card.value;
            const unitHtml = card.unit ? `<span class="card-unit">${this.escapeHtml(card.unit)}</span>` : '';
            const trendHtml = card.trend !== 'neutral'
                ? `<div class="card-trend trend-${card.trend}"></div>`
                : '';
            return `
    <div class="card">
      <div class="card-label">${this.escapeHtml(card.label)}</div>
      <div class="card-value-container">
        <div class="card-value">${this.escapeHtml(String(valueDisplay))}</div>
        ${unitHtml}
      </div>
      ${trendHtml}
    </div>`;
        })
            .join('');
    }
    escapeHtml(text) {
        const map = {
            '&': '&amp;',
            '<': '&lt;',
            '>': '&gt;',
            '"': '&quot;',
            "'": '&#039;',
        };
        return text.replace(/[&<>"']/g, m => map[m]);
    }
}
exports.TemplateRenderer = TemplateRenderer;
