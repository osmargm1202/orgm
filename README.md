
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
sudo curl -L https://github.com/osmargm1202/orgm/releases/latest/download/orgm -o /usr/bin/orgm && sudo chmod +x /usr/bin/orgm
```

## TO-DO

- Portadas (Docs)
- Crear Carpetas
- Crear AI configs
- Cartas
- Calculos 