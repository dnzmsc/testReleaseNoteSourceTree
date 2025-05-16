#![cfg_attr(not(debug_assertions), windows_subsystem = "windows")]

use std::fs::{read_to_string, write};
use std::path::PathBuf;
use tauri::Manager;

#[tauri::command]
fn save_release_note(json_data: String) -> Result<(), String> {
    let mut path = std::env::current_dir().map_err(|e| e.to_string())?;
    path.push("release_notes.json");

    // Legge o crea struttura esistente
    let mut data = serde_json::json!({ "releases": [] });

    if path.exists() {
        let content = read_to_string(&path).map_err(|e| e.to_string())?;
        data = serde_json::from_str(&content).map_err(|e| e.to_string())?;
    }

    let new_note: serde_json::Value = serde_json::from_str(&json_data).map_err(|e| e.to_string())?;
    data["releases"]
        .as_array_mut()
        .ok_or("Formato JSON invalido")?
        .push(new_note);

    write(&path, serde_json::to_string_pretty(&data).unwrap()).map_err(|e| e.to_string())?;

    Ok(())
}

fn main() {
    tauri::Builder::default()
        .invoke_handler(tauri::generate_handler![save_release_note])
        .run(tauri::generate_context!())
        .expect("errore nell'avvio dell'app Tauri");
}
