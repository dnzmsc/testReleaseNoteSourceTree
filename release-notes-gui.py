import tkinter as tk
from tkinter import messagebox
import json
import os
from datetime import date

FILE_PATH = os.path.join(os.getcwd(), "release_notes.json")

def validate_and_save():
    release = {
        "data": date.today().isoformat(),
        "tipo": tipo_var.get(),
        "titolo": titolo_entry.get().strip(),
        "descrizione": descrizione_text.get("1.0", "end").strip(),
        "autore": autore_entry.get().strip(),
        "pr": pr_entry.get().strip(),
        "changelog": changelog_text.get("1.0", "end").strip()
    }

    # Validazioni
    if release["tipo"] not in ["Feature", "Fix", "Refactor"]:
        messagebox.showerror("Errore", "Seleziona un tipo valido.")
        return
    if len(release["titolo"]) < 3:
        messagebox.showerror("Errore", "Il titolo deve avere almeno 3 caratteri.")
        return
    if len(release["descrizione"]) < 10:
        messagebox.showerror("Errore", "Descrizione troppo breve.")
        return
    if not release["pr"].startswith("PR") or not release["pr"][2:].isdigit():
        messagebox.showerror("Errore", "Il PR deve iniziare con 'PR' seguito da numeri.")
        return

    data = {"releases": []}
    if os.path.exists(FILE_PATH):
        try:
            with open(FILE_PATH, "r") as f:
                data = json.load(f)
        except Exception:
            pass

    data["releases"].append(release)
    with open(FILE_PATH, "w") as f:
        json.dump(data, f, indent=2)

    root.destroy()

# GUI
root = tk.Tk()
root.title("Compila Release Notes")

# --- Forza la finestra in primo piano (macOS, Windows, Linux) ---
root.update_idletasks()
root.lift()
root.attributes('-topmost', True)
root.after_idle(lambda: root.attributes('-topmost', False))

# --- Costruzione interfaccia ---
tipo_var = tk.StringVar()
tk.Label(root, text="Tipo (Feature/Fix/Refactor):").pack()
tk.OptionMenu(root, tipo_var, "Feature", "Fix", "Refactor").pack()

tk.Label(root, text="Titolo:").pack()
titolo_entry = tk.Entry(root)
titolo_entry.pack()

tk.Label(root, text="Descrizione:").pack()
descrizione_text = tk.Text(root, height=4)
descrizione_text.pack()

tk.Label(root, text="Autore:").pack()
autore_entry = tk.Entry(root)
autore_entry.pack()

tk.Label(root, text="PR (es: PR1234):").pack()
pr_entry = tk.Entry(root)
pr_entry.pack()

tk.Label(root, text="Changelog:").pack()
changelog_text = tk.Text(root, height=3)
changelog_text.pack()

tk.Button(root, text="Salva", command=validate_and_save).pack(pady=10)

# --- Avvio della GUI ---
root.mainloop()

