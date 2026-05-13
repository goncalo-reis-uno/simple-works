import {describe, it, expect, beforeAll, afterAll} from "vitest"
import axios from "axios"
import {Client} from "pg"


const BASE_URL = "http://localhost:8080"

describe("Auth flow", () => {
  let user
  let refreshToken
  let accessToken
  
  // setup db connection
  const db = new Client({
    host: "localhost",
    port: 7777,
    user: "postgres",
    password: "admin",
    database: "sgc_basic_API"
  })

  // register a test user
  beforeAll(async () => {
    await db.connect()

    user = {
      name: `testuser_${Date.now()}`,
      email: `test${Date.now()}@gmail.com`,
      password: "password123"
    }

    const registerRes = await axios.post(`${BASE_URL}/register`, user)
    expect(registerRes.status).toBe(201)
    
    console.log("Registered test user successfully")
  })

  // test login
  it("should login user", async () => {
    const res = await axios.post(`${BASE_URL}/login`, {
      email: user.email,
      password: user.password
    })

    expect(res.status).toBe(200)
    expect(res.data).toHaveProperty('refresh_token')
    expect(res.data).toHaveProperty('access_token')
    
    refreshToken = res.data.refresh_token
    accessToken = res.data.access_token

    console.log("Logged in test user successfully")
  })

  // test refresh token
  it("should refresh token", async () => {
    const res = await axios.post(`${BASE_URL}/refresh`, {
      refresh_token: refreshToken
    })

    expect(res.status).toBe(200)
    expect(res.data).toHaveProperty('access_token')
    
    console.log("Refreshed token for test user successfully")
  })

  // test logout
  it("should logout user", async () => {
    const res = await axios.post(`${BASE_URL}/logout`, {
      refresh_token: refreshToken
    })

    expect(res.status).toBe(200)
    console.log("Logged out test user successfully")
  })

  // delete the test user and close db connection
  afterAll(async () => {
    await db.query(
      "DELETE FROM users WHERE email = $1",
      [user.email]
    )

    await db.end()
  })
})