import { invoke } from '@tauri-apps/api/core';

document.getElementById("salva").addEventListener("click", async () => {
  console.log ("cliccato");
  const release = {
    data: new Date().toISOString().split('T')[0],
    tipo: document.getElementById("tipo").value,
    titolo: document.getElementById("titolo").value.trim(),
    descrizione: document.getElementById("descrizione").value.trim(),
    autore: document.getElementById("autore").value.trim(),
    pr: document.getElementById("pr").value.trim(),
    changelog: document.getElementById("changelog").value.trim()
  };

  if (!release.tipo || !["Feature", "Fix", "Refactor"].includes(release.tipo)) return alert("Tipo non valido");
  if (!release.titolo || release.titolo.length < 3) return alert("Titolo troppo corto");
  if (!release.descrizione || release.descrizione.length < 10) return alert("Descrizione troppo corta");
  if (!/^PR\d+$/.test(release.pr)) return alert("PR non valido");

  try {
    await invoke("save_release_note", { jsonData: JSON.stringify(release) });
    alert("✅ Release salvata!");
    window.close();
  } catch (e) {
    console.error("Errore:", e);
    alert("Errore nel salvataggio: " + e);
  }
});
