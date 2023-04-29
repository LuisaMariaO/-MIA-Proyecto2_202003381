import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import Service  from "../Services/Service";
import download from 'downloadjs'

function Index() {
    const navigate = useNavigate();
    let fileReader; //Lectura del archivo

    const [file, setFile] = useState(); //Archivo actual que será leído
    const [value, setValue] = useState("");
    const [consolee, setConsolee] = useState("");

    const handleSubmitFile = (event) =>{
        event.preventDefault()
        event.target.reset();
        fileReader = new FileReader()
        fileReader.onloadend = handleFileRead;
        fileReader.readAsText(file)
        

    };

    const handleFileRead =(e) =>{
        const content = fileReader.result
       // alert(content)
        setValue(content)
        //miEditor.current.editor.setValue(content)
        //miEditor.current.editor.clearSelection()
    };

    const handleChangeFile = (event) => {
        setFile(event.target.files[0])
       
    };

    const handleChangeValue = (event) => {
        setValue(event.target.value);
        // alert(value)
      };

    const handleSubmitComand = (event) => {
      event.preventDefault();
      setConsolee("Cargando...")
      Service.postCode(value)
      .then(({result}) => {
        
        setConsolee(result);
        console.log(result)
      
    });

    
    
    };

    const handleClean = (event) =>{
      event.preventDefault();
      setValue("")
    };

    const handleLogout = (event)=>{
      event.preventDefault();
      Service.logout()
      .then(({status, message}) => {
        if(status!=null){
          if (status==48){
            alert(message)
          }else{
            alert("¡Sesión cerrada!")
          }
        }else{
          alert("Ocurrió un error de conexión")
        }
      
    });
    };

    const handleReports = (event) => {
      event.preventDefault()
      Service.reports()
      .then(({reportes}) => {
        if (reportes!=null){
          for (let i=0; i<reportes.length;i++){
            download(atob(reportes[i].reporte),reportes[i].nombre, { type: "image/jpeg"  || "text/plain"});
          }
        }
      
    });
    };
  return (
    <>
    <nav class="navbar bg-dark" data-bs-theme="dark">
    <div class="container-fluid">
      <a class="navbar-brand" onClick={ ()=>navigate('/') } href="">
        <img src="https://cdn-icons-png.flaticon.com/512/3767/3767084.png" alt="Logo" width="30" height="24" class="d-inline-block align-text-top"/>
        MIA Proyecto2
      </a>
      &nbsp; &nbsp; &nbsp;
      <ul class="navbar-nav me-auto mb-2 mb-lg-0">

      <li class="nav-item">
          <a class="nav-link"  href="" onClick={handleReports}>Reportes</a>
        </li>
        </ul>
      <div class="btn-group" role="group" aria-label="Basic example">
      <button class="btn btn-outline-success" type="button" onClick={()=>navigate('/login')}>Login</button>
      <button class="btn btn-outline-danger" type="button" onClick={ handleLogout }>Logout</button>
      </div>
    </div>
  </nav>
   <br></br> 
   <div class="px-5">
  <form onSubmit={handleSubmitFile}>
          <input type="file" accept=".eea" onChange={handleChangeFile}  />
          <button type="submit" class="btn btn-warning">Cargar</button>
  </form>
  </div>
<br></br>
  
  <div class="mb-3 px-5">
  <form onSubmit={handleSubmitComand }>
  <textarea class="form-control" id="editor" rows="10" value={ value } onChange={handleChangeValue}></textarea>
  <button type="submit" class="btn btn-success">
  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-play-fill" viewBox="0 0 16 16">
  <path d="m11.596 8.697-6.363 3.692c-.54.313-1.233-.066-1.233-.697V4.308c0-.63.692-1.01 1.233-.696l6.363 3.692a.802.802 0 0 1 0 1.393z"/>
  </svg> 
  Ejecutar
  </button>
    &nbsp;
  <button type="button" class="btn btn-secondary" onClick={ handleClean }>
  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" fill="currentColor" class="bi bi-stars" viewBox="0 0 16 16">
  <path d="M7.657 6.247c.11-.33.576-.33.686 0l.645 1.937a2.89 2.89 0 0 0 1.829 1.828l1.936.645c.33.11.33.576 0 .686l-1.937.645a2.89 2.89 0 0 0-1.828 1.829l-.645 1.936a.361.361 0 0 1-.686 0l-.645-1.937a2.89 2.89 0 0 0-1.828-1.828l-1.937-.645a.361.361 0 0 1 0-.686l1.937-.645a2.89 2.89 0 0 0 1.828-1.828l.645-1.937zM3.794 1.148a.217.217 0 0 1 .412 0l.387 1.162c.173.518.579.924 1.097 1.097l1.162.387a.217.217 0 0 1 0 .412l-1.162.387A1.734 1.734 0 0 0 4.593 5.69l-.387 1.162a.217.217 0 0 1-.412 0L3.407 5.69A1.734 1.734 0 0 0 2.31 4.593l-1.162-.387a.217.217 0 0 1 0-.412l1.162-.387A1.734 1.734 0 0 0 3.407 2.31l.387-1.162zM10.863.099a.145.145 0 0 1 .274 0l.258.774c.115.346.386.617.732.732l.774.258a.145.145 0 0 1 0 .274l-.774.258a1.156 1.156 0 0 0-.732.732l-.258.774a.145.145 0 0 1-.274 0l-.258-.774a1.156 1.156 0 0 0-.732-.732L9.1 2.137a.145.145 0 0 1 0-.274l.774-.258c.346-.115.617-.386.732-.732L10.863.1z"/>
  </svg>
  Limpiar
  </button>
  </form>
  
 </div>

 <div class="mb3 px-5">
 <textarea class="form-control bg-dark" id="editor" rows="10" value={ consolee } readOnly={true} style={{color:"#FFFFFF"}}></textarea>
 </div>

  </>
  );
}

export default Index;