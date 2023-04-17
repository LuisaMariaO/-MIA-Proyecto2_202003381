import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import Service  from "../Services/Service";

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
      Service.postCode(value)
      .then(({result}) => {
        
        setConsolee(result);
        console.log(result)
      
    });
    
    };
  return (
    <>
    <nav class="navbar bg-dark" data-bs-theme="dark">
    <div class="container-fluid">
      <a class="navbar-brand" onClick={ ()=>navigate('/') } href="/">
        <img src="https://cdn-icons-png.flaticon.com/512/3767/3767084.png" alt="Logo" width="30" height="24" class="d-inline-block align-text-top"/>
        MIA Proyecto2
      </a>
      <div class="btn-group" role="group" aria-label="Basic example">
      <button class="btn btn-outline-success" type="button">Login</button>
      <button class="btn btn-outline-danger" type="button">Logout</button>
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
  </form>
  
 </div>

 <div class="mb3 px-5">
 <textarea class="form-control bg-dark" id="editor" rows="10" value={ consolee } readOnly={true} style={{color:"#FFFFFF"}}></textarea>
 </div>

  </>
  );
}

export default Index;