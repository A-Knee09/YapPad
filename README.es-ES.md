# YapPad

Una aplicación de toma de notas basada en terminal construida con Go y Bubble Tea.

> Construido mientras aprendía el desarrollo de TUIs con [BubbleTea](https://github.com/charmbracelet/bubbletea)

## Showcase

https://github.com/user-attachments/assets/cbe425a2-4a54-4b5e-a331-47501feb353d

<img width="1920" height="1080" alt="swappy-20260701_215440" src="https://github.com/user-attachments/assets/48a30f40-ca7b-422a-ae85-3da3626031bb" />



<img width="1920" height="1080" alt="260326_00h48m57s_screenshot" src="https://github.com/user-attachments/assets/82eb537b-716f-49ae-bf9b-c334eb67b2b0" />

> [!IMPORTANT]
> - Probado únicamente en Linux y MacOS
> - La vista previa de imágenes requiere una terminal compatible con Kitty (Kitty, WezTerm) y tener `chafa` instalado
> - La vista previa de imágenes puede tardar un tiempo en cargar inicialmente o al redimensionar la terminal, pero no volverá a ocurrir

## Requisitos

- Go 1.21+
- [chafa](https://hpjansson.org/chafa/) (opcional, para vistas previas de imágenes)

## Instalación

```bash
git clone https://github.com/A-Knee09/YapPad.git
cd YapPad
make install
```

Se instala en `~/.local/bin/yap`. Asegúrate de que esté en tu PATH:

```bash
# Bash/Zsh
export PATH="$HOME/.local/bin:$PATH"

# Fish
fish_add_path ~/.local/bin
```

```bash
make uninstall  # para eliminar
```

## Primera Ejecución

En la primera ejecución, YapPad te pedirá configurar:

- Dónde almacenar tus notas (predeterminado: `~/.YapPad`)
- Qué editor utilizar (predeterminado: `inbuilt`)
- Qué tema utilizar (predeterminado: `default`)

La configuración se guarda en `~/.config/yappad/config.toml` y puede editarse directamente.

## Uso

```bash
yap                   # abre tu bóveda configurada
yap .                 # abre el directorio actual como bóveda (solo para esta sesión)
yap --theme gruvbox   # sobrescribe el tema para esta sesión
yap --editor nvim     # sobrescribe el editor para esta sesión
yap --version         # imprime la versión
yap --help            # muestra la ayuda
```

## Características

### Notas y Archivos

Presiona `ctrl n` para crear una nota nueva. Se te pedirá un nombre de archivo y luego una descripción opcional. Presionar enter con el nombre de archivo vacío crea un archivo con la fecha actual. Presionar enter con la descripción vacía la omite y muestra la fecha de modificación por defecto.

Cada nota puede tener una descripción personalizada que se muestra debajo de su título en la lista. Las descripciones se almacenan en `.metadesc/` como archivos adjuntos (sidecar) y nunca modifican el contenido de la nota.

Presiona `ctrl r` para renombrar una nota. Se te pedirá el nuevo nombre y la descripción. Omitir la descripción conservará la existente.


### Temas
<table>
  <tr>
    <td><img src="https://github.com/user-attachments/assets/cf25288b-df85-4d16-92d1-8901aacac6d3" alt="Default" width="450"/></td>
    <td><img src="https://github.com/user-attachments/assets/b32972f7-2489-4ab1-9bf8-f23506ce01b4" alt="Dracula" width="450"/></td>
  </tr>
  <tr>
    <td align="center">Default</td>
    <td align="center">Dracula</td>
  </tr>
</table>
<table>
  <tr>
    <td><img src="https://github.com/user-attachments/assets/d1445109-5337-4dbd-949f-6f1f1ca83e2c" alt="Dusk" width="450"/></td>
    <td><img src="https://github.com/user-attachments/assets/8aa1c819-3deb-4d0c-858b-c75a0f01865c" alt="Forest" width="450"/></td>
  </tr>
  <tr>
    <td align="center">Dusk</td>
    <td align="center">Forest</td>
  </tr>
</table>
<table>
  <tr>
    <td><img src="https://github.com/user-attachments/assets/d6fa6c25-0c28-4d04-bcd4-fd466e6d3b6a" alt="Glacier" width="450"/></td>
    <td><img src="https://github.com/user-attachments/assets/7a1b0049-8f8b-4cea-83e2-3f4a6e66a2ff" alt="Gruvbox" width="450"/></td>
  </tr>
  <tr>
    <td align="center">Glacier</td>
    <td align="center">Gruvbox</td>
  </tr>
</table>
<table>
  <tr>
    <td><img src="https://github.com/user-attachments/assets/1e25cae9-64f8-41ba-b765-ea23bafe4c3e" alt="Moss" width="450"/></td>
    <td><img src="https://github.com/user-attachments/assets/a9893933-8962-4c21-902e-cec55d77eaa1" alt="Nord" width="450"/></td>
  </tr>
  <tr>
    <td align="center">Moss</td>
    <td align="center">Nord</td>
  </tr>
</table>
<table>
  <tr>
    <td><img src="https://github.com/user-attachments/assets/c0d0328c-144d-4634-af23-f5a10d882b9d" alt="Plum" width="450"/></td>
    <td><img src="https://github.com/user-attachments/assets/5cd6360e-f2ff-4aa5-89d5-a0190d9434c2" alt="Solarized" width="450"/></td>
  </tr>
  <tr>
    <td align="center">Plum</td>
    <td align="center">Solarized</td>
  </tr>
</table>
<table>
  <tr>
    <td><img src="https://github.com/user-attachments/assets/67e5a2c8-3037-40a7-ab8c-dd98a39d0461" alt="Tide" width="450"/></td>
    <td><img src="https://github.com/user-attachments/assets/ffec77c5-e23c-4335-8262-085fbad83276" alt="Tokyonight" width="450"/></td>
  </tr>
  <tr>
    <td align="center">Tide</td>
    <td align="center">Tokyonight</td>
  </tr>
</table>

16 temas integrados: `default`, `gruvbox`, `nord`, `tokyonight`, `forest`, `solarized`, `dracula`, `dusk`, `tide`, `moss`, `glacier`, `plum`, `algae`, `sunny`, `stone`.

Configúralos en el archivo de configuración o sobrescríbelos por sesión con `--theme <nombre>`.

### Panel de Vista Previa

Actívalo/desactívalo con `ctrl+p`. Muestra texto con resaltado de sintaxis, vistas previas de markdown y vistas previas de imágenes integradas para formatos compatibles. Se oculta automáticamente si la terminal es demasiado estrecha. La vista previa de imágenes requiere `chafa` y una terminal compatible con Kitty.


### Ordenamiento

Presiona `ctrl+s` para alternar entre los modos de ordenamiento: Modificado (más nuevo/más viejo), Creado (más nuevo/más viejo), Alfabético (ascendente/descendente).

### Editores

Configura `editor` en el archivo de configuración o usa la bandera `--editor`. Soporta `inbuilt`, `nano`, `nvim`, `vim`, `hx`, o cualquier editor que esté en tu `$PATH`. El editor integrado soporta `ctrl+s` para guardar y `ctrl+q` para cerrar.

### Soporte de Mouse

Desplázate por la lista de archivos y el panel de vista previa de forma independiente con la rueda del ratón.

## Atajos de Teclado

| Tecla | Acción |
|-----|--------|
| `ctrl n` | Nueva nota |
| `ctrl r` | Renombrar nota |
| `ctrl d` | Eliminar nota |
| `enter` | Abrir en editor |
| `ctrl+p` | Alternar vista previa |
| `ctrl+s` | Alternar ordenamiento |
| `/` | Filtrar notas |
| `?` | Alternar ayuda |
| `esc` | Cancelar |
| `q` | Salir |

## Configuración

`~/.config/yappad/config.toml`:

```toml
theme = "default"
editor = "inbuilt"
vault = "/home/user/.YapPad"
```

## Almacenamiento

```
~/.YapPad/
├── .metadesc/     # descripciones de notas
└── .templates/    # plantillas de notas opcionales
```

Las notas son archivos planos. Se soporta cualquier tipo de archivo: markdown, texto, imágenes, archivos de código.

## Desarrollo

```bash
make run     # ejecutar directamente
make build   # compilar en build/yap
make clean   # eliminar artefactos de compilación
```
