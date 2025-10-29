import axios from "axios";

const api = axios.create({
    baseURL: "http://localhost:8080",
    withCredentials: true, // Important: Send cookies with requests
    headers: {
        "Content-Type": "application/json",
    },
});

export default api;