import React, { useState } from "react";
import { useNavigate } from "react-router-dom";

function Index() {
    const navigate = useNavigate();
    let fileReader; //Lectura del archivo

    const [file, setFile] = useState(); //Archivo actual que será leído
    const [value, setValue] = useState("");

    const handleSubmitFile = (event) =>{
        event.preventDefault()
        event.target.reset();
        fileReader = new FileReader()
        fileReader.onloadend = handleFileRead;
        fileReader.readAsText(file)
        

    };

    const handleFileRead =(e) =>{
        const content = fileReader.result
      //  alert(content)
        setValue(content)
        //miEditor.current.editor.setValue(content)
        //miEditor.current.editor.clearSelection()
    };

    const handleChangeFile = (event) => {
        setFile(event.target.files[0])
       
    }

    const handleChangeValue = (event) => {
        setValue(event.target.value);
        // alert(value)
      }
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
  <form onSubmit={handleSubmitFile}>
          <input type="file" accept=".eea" onChange={handleChangeFile}  />
          <button type="submit" class="btn btn-warning">Cargar</button>
  </form>
<br></br>
  <div class="mb-3 px-5">
  
  <textarea class="form-control" id="editor" rows="10" defaultValue={ value } onChange={handleChangeValue}></textarea>
 </div>
  </>
  );
}

export default Index;