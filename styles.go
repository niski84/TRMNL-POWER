package main

const tailwindCSS = `    * {
      margin: 0;
      padding: 0;
      box-sizing: border-box;
    }
    
    body {
      width: 800px;
      height: 480px;
      background: #ffffff;
      color: #000000;
      font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Arial, sans-serif;
      overflow: hidden;
      display: flex;
      flex-direction: column;
    }
    
    .header {
      height: 50px;
      min-height: 50px;
      max-height: 50px;
      background: #000000;
      color: #ffffff;
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 0 12px;
      border-bottom: 2px solid #000000;
      flex-shrink: 0;
    }
    
    .header-title {
      font-size: 20px;
      font-weight: bold;
      letter-spacing: -0.2px;
      line-height: 1.2;
    }
    
    .header-timestamp {
      font-size: 14px;
      font-weight: 500;
      line-height: 1.2;
    }
    
    .content {
      flex: 1;
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 10px;
      padding: 10px;
      min-height: 0;
      overflow: hidden;
      max-height: 420px;
    }
    
    .dashboard-grid {
      display: contents;
    }
    
    .card {
      background: #ffffff;
      border: 2px solid #000000;
      border-radius: 0;
      padding: 8px 10px;
      display: flex;
      flex-direction: column;
      justify-content: center;
      min-height: 0;
      overflow: hidden;
      box-sizing: border-box;
    }
    
    .card-label {
      font-size: 12px;
      font-weight: bold;
      color: #000000;
      margin-bottom: 4px;
      text-transform: uppercase;
      letter-spacing: 0.3px;
      line-height: 1.1;
      flex-shrink: 0;
    }
    
    .card-value-container {
      display: flex;
      align-items: baseline;
      gap: 3px;
      flex-wrap: nowrap;
      min-height: 0;
      align-items: flex-end;
    }
    
    .card-value {
      font-size: 32px;
      font-weight: bold;
      color: #000000;
      line-height: 1;
      flex-shrink: 1;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    
    .card-unit {
      font-size: 16px;
      font-weight: 600;
      color: #000000;
      flex-shrink: 0;
      line-height: 1;
      padding-bottom: 2px;
    }
    
    .card-completed {
      opacity: 0.6;
    }
    
    .card-completed .card-label {
      text-decoration: line-through;
    }
    
    .card-trend {
      margin-top: 8px;
      font-size: 18px;
      font-weight: 500;
    }
    
    .trend-up::before {
      content: '▲ ';
    }
    
    .trend-down::before {
      content: '▼ ';
    }
    
    .trend-neutral::before {
      content: '● ';
    }
    
    .content.single {
      grid-template-columns: 1fr;
    }
    
    .content.three .card:first-child {
      grid-column: span 2;
    }
    
    /* Todo list styles */
    .todo-container {
      width: 100%;
      padding: 20px;
    }
    
    .todo-header {
      font-size: 28px;
      font-weight: bold;
      margin-bottom: 20px;
      color: #000000;
      border-bottom: 3px solid #000000;
      padding-bottom: 10px;
    }
    
    .todo-list {
      list-style: none;
      padding: 0;
      margin: 0;
    }
    
    .todo-item {
      display: flex;
      align-items: center;
      padding: 12px 0;
      border-bottom: 2px solid #cccccc;
      font-size: 24px;
      gap: 12px;
    }
    
    .todo-item:last-child {
      border-bottom: none;
    }
    
    .todo-item.completed {
      opacity: 0.6;
    }
    
    .todo-checkbox {
      font-size: 28px;
      font-weight: bold;
      width: 40px;
      text-align: center;
      color: #000000;
    }
    
    .todo-text {
      flex: 1;
      color: #000000;
      font-weight: 500;
    }
    
    .todo-item.completed .todo-text {
      text-decoration: line-through;
    }
    
    .todo-category {
      font-size: 18px;
      padding: 4px 12px;
      background: #000000;
      color: #ffffff;
      border: 2px solid #000000;
      font-weight: bold;
    }`

