import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import Service  from "../Services/Service";

function Login() {
    const navigate = useNavigate();
    

    const [id, setId] = useState(""); 
    const [user, setUser] = useState("");
    const [password, setPassword] = useState("");



    const handleChangeId = (event) => {
        setId(event.target.value)
    };
    const handleChageUser = (event) => {
      setUser(event.target.value) 
    };
  const handleChangePassword = (event) => {
    setPassword(event.target.value)
    };

    const handleSubmit = (event) => {
      event.preventDefault();
      //setConsolee("Cargando...")
      Service.login(id,user,password)
      .then(({status,message}) => {
        if(status!=null){
        if (status==48){
          alert(message)
          navigate('/')
        }else{
          alert("¡Sesión iniciada!")
          navigate('/')
        }
      }else{
        alert("Ocurrió un error de comunicación :(")
      }
      
    });

    
    
    };

  
  return (
    <>
    <div class="container">
        <div class="row py-5">
            <h3>Login</h3>
            <div class="col px-5 py-5 bg-success border border-success-subtle rounded-3" >
            <form onSubmit={handleSubmit}>
  <div class="mb-3">
    <label for="id" class="form-label">ID Partición</label>
    <input type="text" class="form-control" id="id" onChange={handleChangeId} required={true}/>
  </div>
  <div class="mb-3">
    <label for="user" class="form-label">User</label>
    <input type="text" class="form-control" id="user" required onChange={handleChageUser}/>
  </div>
  <div class="mb-3">
    <label for="password" class="form-label">Password</label>
    <input type="password" class="form-control" id="password" required onChange={handleChangePassword}/>
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