const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

function run() {
    const extensionDir = path.dirname(__dirname);
    console.log(`Packaging VS Code extension from directory: ${extensionDir}`);

    try {
        console.log("Installing VS Code Extension Manager (vsce) locally...");
        // Using npx is cleaner because it does not pollute the global environment
        console.log("Running vsce packaging command...");
        execSync("npx -y @vscode/vsce package", {
            cwd: extensionDir,
            stdio: 'inherit'
        });
        console.log("VS Code extension packaged successfully!");
        
        // Find packaged vsix file
        const files = fs.readdirSync(extensionDir);
        const vsixFile = files.find(f => f.endsWith('.vsix'));
        if (vsixFile) {
            console.log(`Package file created: ${path.join(extensionDir, vsixFile)}`);
            console.log("You can install it in VS Code via: code --install-extension " + vsixFile);
        }
    } catch (error) {
        console.error("Packaging failed:", error.message);
        process.exit(1);
    }
}

run();
