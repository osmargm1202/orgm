
## WINDOWS

### Download

```
New-Item -Path "$env:USERPROFILE\.config\orgm" -ItemType Directory -Force

powershell -Command "Invoke-WebRequest -Uri https://github.com/osmargm1202/orgm/releases/latest/download/orgm.exe -OutFile '$env:USERPROFILE\.config\orgm\orgm.exe'"
```

### set path

```
setx PATH "%PATH%;$env:USERPROFILE\.config\orgm"
```


## LINUX

### Download

```
curl -L https://github.com/osmargm1202/orgm/releases/latest/download/orgm -o $HOME/.local/bin/orgm && chmod +x $HOME/.local/bin/orgm
```

## TO-DO

- Portadas (Docs)
- Crear Carpetas
- Crear AI configs
- Buscar empresa RNC
- Cartas
- Calculos
- 
