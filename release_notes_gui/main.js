const { app, BrowserWindow, ipcMain } = require('electron');
const fs = require('fs');
const path = require('path');

function createWindow() {
  const win = new BrowserWindow({
    width: 500,
    height: 400,
    resizable: false,
    webPreferences: {
      nodeIntegration: true,
      contextIsolation: false
    }
  });

  win.loadFile('index.html');
}

ipcMain.on('save-release-notes', (event, data) => {
  const targetPath = path.resolve(process.cwd(), 'release_notes.json');
  fs.writeFileSync(targetPath, JSON.stringify(data, null, 2));
  app.quit();
});

app.whenReady().then(() => {
  createWindow();
});
