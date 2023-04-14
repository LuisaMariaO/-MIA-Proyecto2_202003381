import axios from 'axios'

const instance = axios.create(
    {
        baseURL: 'http://localhost:8000',
        timeout: 15000,
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