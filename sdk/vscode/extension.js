const vscode = require('vscode');

/**
 * @param {vscode.ExtensionContext} context
 */
function activate(context) {
    console.log('Lychee extension is now active!');

    let disposable = vscode.commands.registerCommand('lychee.runModel', async function () {
        const model = await vscode.window.showInputBox({ prompt: 'Enter model name to run (e.g., llama3)' });
        if (model) {
            const terminal = vscode.window.createTerminal(`Lychee: ${model}`);
            terminal.show();
            terminal.sendText(`lychee run ${model}`);
        }
    });

    context.subscriptions.push(disposable);
}

function deactivate() {}

module.exports = {
    activate,
    deactivate
}
