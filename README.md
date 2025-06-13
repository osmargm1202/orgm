
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

```
setx PATH "%PATH%;$env:USERPROFILE\Nextcloud\Apps\bin"
```


## LINUX

### Download

```
sudo curl -L https://github.com/osmargm1202/orgm/releases/latest/download/orgm -o /usr/bin/orgm && sudo chmod +x /usr/bin/orgm
```

### set path

```
echo 'export PATH=$PATH:$HOME/Nextcloud/Apps/bin' >> ~/.bashrc
source ~/.bashrc

echo 'export PATH=$PATH:$HOME/Nextcloud/Apps/bin' >> ~/.zshrc
source ~/.zshrc

echo 'export PATH=$PATH:$HOME/Nextcloud/Apps/bin' >> ~/.profile
source ~/.profile

echo 'export PATH=$PATH:$HOME/Nextcloud/Apps/bin' >> ~/.bash_profile
source ~/.bash_profile

```

## TO-DO

- Portadas (Docs)
- Crear AI configs
- guardar conversacion
- Cartas
- Calculos
- buscar cotizaciones
- Schemas y configs en nube y RNC en Nextcloud