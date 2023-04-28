import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import Service  from "../Services/Service";

function Login() {
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
  return (
    <>
    <div class="container">
        <div class="row py-5">
            <h3>Login</h3>
            <div class="col px-5 py-5 bg-success border border-success-subtle rounded-3" >
            <form>
  <div class="mb-3">
    <label for="id" class="form-label">ID Partición</label>
    <input type="text" class="form-control" id="id" required/>
  </div>
  <div class="mb-3">
    <label for="user" class="form-label">User</label>
    <input type="text" class="form-control" id="user" required/>
  </div>
  <div class="mb-3">
    <label for="password" class="form-label">Password</label>
    <input type="password" class="form-control" id="password" required/>
  </div>
  <button type="submit" class="btn btn-warning">Ingresar</button>
</form>
            </div>
        </div>
    </div>

   </>
  );
}

export default Login;