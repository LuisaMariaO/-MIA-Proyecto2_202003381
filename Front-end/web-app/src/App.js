import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import './App.css';
import Index from "./Pages";
import '../node_modules/bootstrap/dist/css/bootstrap.min.css';
import Login from "./Pages/login";

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<Index/>}></Route>
        <Route path="/login" element={<Login></Login>}></Route>
        <Route path="*" element={<Navigate to="/" replace={true}></Navigate>} exact={true}></Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
