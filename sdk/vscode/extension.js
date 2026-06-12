const vscode = require('vscode');
const http = require('http');

function request(method, path, body) {
    return new Promise((resolve, reject) => {
        const req = http.request({
            hostname: 'localhost',
            port: 11434,
            path: path,
            method: method,
            headers: { 'Content-Type': 'application/json' }
        }, (res) => {
            let data = '';
            res.on('data', (chunk) => data += chunk);
            res.on('end', () => {
                if (res.statusCode >= 400) {
                    reject(new Error(data || `HTTP ${res.statusCode}`));
                } else { resolve(data); }
            });
        });
        req.on('error', reject);
        if (body) { req.write(JSON.stringify(body)); }
        req.end();
    });
}

function activate(context) {
    console.log('Lychee extension is now active!');

    let runModelDisposable = vscode.commands.registerCommand('lychee.runModel', async function () {
        const model = await vscode.window.showInputBox({ prompt: 'Enter model name to run (e.g., llama3)' });
        if (model) {
            const terminal = vscode.window.createTerminal(`Lychee: ${model}`);
            terminal.show();
            terminal.sendText(`lychee run ${model}`);
        }
    });
    context.subscriptions.push(runModelDisposable);

    let listModelsDisposable = vscode.commands.registerCommand('lychee.listModels', async function () {
        try {
            const res = await request('GET', '/api/tags');
            const data = JSON.parse(res);
            const models = data.models || [];
            if (models.length === 0) {
                vscode.window.showInformationMessage('No models found.');
                return;
            }
            const selection = await vscode.window.showQuickPick(models.map(m => m.name), {
                placeHolder: 'Select a model to view details'
            });
            if (selection) {
                const selectedModel = models.find(m => m.name === selection);
                vscode.window.showInformationMessage(
                    `Model: ${selectedModel.name} | Size: ${(selectedModel.size / (1024*1024*1024)).toFixed(2)} GB | Digest: ${selectedModel.digest.substring(0, 12)}`
                );
            }
        } catch (err) {
            vscode.window.showErrorMessage(`Failed to list models: ${err.message}`);
        }
    });
    context.subscriptions.push(listModelsDisposable);

    let structuredDisposable = vscode.commands.registerCommand('lychee.structured', async function () {
        try {
            const resTags = await request('GET', '/api/tags');
            const modelNames = (JSON.parse(resTags).models || []).map(m => m.name);
            if (modelNames.length === 0) {
                vscode.window.showErrorMessage('No models available.');
                return;
            }
            const model = await vscode.window.showQuickPick(modelNames, { placeHolder: 'Select Model' });
            if (!model) return;

            const prompt = await vscode.window.showInputBox({ prompt: 'Enter prompt' });
            if (!prompt) return;

            const schemaStr = await vscode.window.showInputBox({
                prompt: 'Enter JSON Schema object string',
                value: '{"type": "object", "properties": {"summary": {"type": "string"}}, "required": ["summary"]}'
            });
            if (!schemaStr) return;
            const schema = JSON.parse(schemaStr);

            vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: "Generating structured output...",
                cancellable: false
            }, async () => {
                const res = await request('POST', '/api/structured', { model, prompt, schema });
                const result = JSON.parse(res);
                const doc = await vscode.workspace.openTextDocument({
                    content: JSON.stringify(JSON.parse(result.output), null, 2),
                    language: 'json'
                });
                await vscode.window.showTextDocument(doc);
            });
        } catch (err) {
            vscode.window.showErrorMessage(`Structured generation failed: ${err.message}`);
        }
    });
    context.subscriptions.push(structuredDisposable);

    let composeDisposable = vscode.commands.registerCommand('lychee.compose', async function () {
        try {
            const input = await vscode.window.showInputBox({ prompt: 'Enter input text' });
            if (!input) return;

            const stepsStr = await vscode.window.showInputBox({
                prompt: 'Enter Compose Steps (JSON array)',
                value: '[{"model": "gemma3", "prompt": "Sentiment: {{input}}"}]'
            });
            if (!stepsStr) return;
            const steps = JSON.parse(stepsStr);

            vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: "Executing compose pipeline...",
                cancellable: false
            }, async () => {
                const res = await request('POST', '/api/compose', { input, steps });
                const result = JSON.parse(res);
                const doc = await vscode.workspace.openTextDocument({
                    content: JSON.stringify(result, null, 2),
                    language: 'json'
                });
                await vscode.window.showTextDocument(doc);
            });
        } catch (err) {
            vscode.window.showErrorMessage(`Composition failed: ${err.message}`);
        }
    });
    context.subscriptions.push(composeDisposable);
}

function deactivate() {}

module.exports = {
    activate,
    deactivate
}
