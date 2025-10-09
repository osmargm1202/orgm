# admAPP

En este proyecto vamos a realizar un frontend en wails. el comando esta disponible wails dev. 

proyecto en wails, react-ts.

## descripcion del programa

arriba el programa tendra un footer con el logo de la empresa que esta en public orgm.png nombre de la aplicacion admAPP y al final 
el frontend debe consumir las apis de @EJEMPLOS_API.md 

toda la logica de cosumir la api en main.go

se rata de un dashboard adminitrativo con los siguientes menus. debe tener un menu a la izquierda que se quitan los nombres si se ponen pequenos, utiliza bootstrap y iconos de bootstrap para los estilos. trata de que sea darkmode con react. el fashboard debe poder hacer todas las funciones descritas en la api.

para los botones me gustan estos:

utiliza darkmode tipo tokyo night

/* Buttons */
.btn-primary {
    background: linear-gradient(135deg, #0044cc, #0088ff);
    color: #ffffff;
    border: none;
    padding: 12px 24px;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: transform 0.2s ease, box-shadow 0.3s ease, background 0.3s ease;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    text-decoration: none;
    box-shadow: 0 0 10px rgba(0, 150, 255, 0.6);
}

.btn-primary:hover {
    transform: translateY(-2px);
    box-shadow: 0 0 15px rgba(0, 150, 255, 0.9);
    background: linear-gradient(135deg, #0055ff, #00aaff);
}

.btn-secondary {
    background: transparent;
    color: #cfe9ff;
    border: 1px solid rgba(0, 150, 255, 0.4);
    padding: 12px 24px;
    border-radius: 8px;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.3s ease;
    display: inline-flex;
    align-items: center;
    gap: 8px;
    text-decoration: none;
}

.btn-secondary:hover {
    background: rgba(0, 136, 255, 0.1);
    border-color: rgba(0, 150, 255, 0.8);
    transform: translateY(-2px);
}


los iconos utiilza los de bootstrap.

vamos a desarrollar el proyecto en fases y vamos hacer un stop en cada fase hasta verificar que funciona correctamente.

el menu de la izquierda tendra:

## fases

- clientes
- proyectos
- Cotizaciones
- pagos
- estados
- facturas
- comprobantes

## clientes

en clientes tendre una barra de busqueda segudo de un boton que dice nuevo.
debajo habra una lista con todos los clientes, que se filtratara por nombre, rnc, nombre comercial.
al selecconar un cliente tengo la opcion que debajo se carguen campos con los datos actuales y pueda modificarlos o subir un logo del cliente que dbee mostrasre is existe.
el boton nuevo activa los mismo campos pero con la premiza de cuando le de a guardar se guarde como un cliente nuevo. debe haber una casilla donde si pongo el id del cliente se selecciona tambien, en la tabla solo aparece id, nombre, rnc, nombre, nombre comercial, representante, 

## proyectos

la ficha de proyectos tienee la misma barraa de cliente y lista de clientes para filtrar, pero cuando selecciono un cliente me permite ver los proyectos que tiene, cuando selecciono un proyecto me permite modificar los cambpos, abra un boton que aparce luego de seleccionar un cliente donde puedo agregar proyectos y cuando selecciono un proyecto abra otro boton que aparece que me permite crear una cotizacion a ese proyecto. debe haber una casilla donde si pongo el id del proyecto se selecciona tambien, en la tabla solo debe aparecer id, nombre, ubicacion.

al crear una cotizacion se crea con los datos de cliente proyecto y condatos por defecto, como la fecha, descripcion vacia, tiempo de entrega de 30 dias, tasa 1, moneda RD$, avance de 60, estado "GENERADA", idioma ES, retencion NINGUNA, descuentop 0., retencionp 0, itbisp 0.

al crear esta cotizacion tengo la opcion de darle al boton para empezar a editar estos valores o crear mas cotizzaciones del proyecto. si presiono editar debe enviarme a la pestana cotizaciones con la cotizacion seleccionada.

## cotizaciones

la ficha cotizaciones comienza con una barra de busqueda de las cotizaciones existtente, una tabla de cotizaciones con las ultimas 10 y un boton de crear nueva, y casila para poner el id de la cotizacion

al seleccionar una cotizacion se activan unos campos donde puedo editar los datos de la cotizacion, adicional a eso tengo los bootens de imprimir en pdf, y un cadro donde se visualizan los pagos asignados que tiene la cotizacion y un porcentaje del monto total.

tambien carga de presupuesto el total del presupuesto de la cotizacion tanto en indirectos como en subtotal y calcula los totales del presupuesto y los presenta, al cambiar el descuento, itbis, retencion los totales se actualizan.

antes de imprimir se debe guardar pedir al usuario guardar si no ha guardado y ha modificado la cotizacion.


## pagos

en esta pestana esta una barra de busqueda de cientes, con casilla de id, donde cuando seeecciono el cliente en otra taba se pueden ver los pagos recibidos de cliente,, cada pago tiene la posibiidad de subir un comprobante. en pdf o imagen. tambien al seleccionar un cliente hay un boton de nuevo pago donde puedo insertar e lpago en USD o RD$ y la fecha, luego a la derecha de los datos de pago, hay una lista de cotizaciones que estan pendientes de pago con su id, proyecto y cuanto falta para saldar, entonces hay una casilla donde puedo poner el monto o en otra casilla el porrcentaje del pago que se asignara a esta cotizacion y me muestra en otra casila cuanto pago tiene total y otra casilla cuanto fata para el 100%.

mas abajo debe haber una tabla con los pagos que faltan por 

## 

