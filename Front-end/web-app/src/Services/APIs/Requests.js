import axios from 'axios'

const instance = axios.create(
    {
        baseURL: 'http://localhost:8000', //la ip podría cambiar http://44.202.167.5
        timeout: 100000,
        headers:{
            'Content-Type':'application/json'
        }
    }
)

export const test = async(value) =>{
    const { data } = await instance.get("/info", { peticion: value })
    return data
}

export const postCode = async(value) =>{
    const { data } = await instance.post("/postCode", { comando: value })
    return data
}

export const login = async(idr, userr, passwordr) =>{
    const { data } = await instance.post("/login", { id:idr, user:userr, password:passwordr })
    return data
}

export const logout = async() =>{
    const { data } = await instance.post("/logout")
    return data
}

export const reports = async() =>{
    const { data } = await instance.get("/reports")
    return data
}
