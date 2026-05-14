import { describe, it, expect, beforeAll, afterAll } from "vitest"
import axios from "axios"
import { Client } from "pg"

const BASE_URL = "http://localhost:8080"

const db = new Client({
  host: "localhost",
  port: 7777,
  user: "postgres",
  password: "admin",
  database: "sgc_basic_API",
})

describe("Tests to verify user routes", () => {
    let userToken
    let adminToken

    let userEmail
    let testUserId
    let adminUserId

    beforeAll(async () => {
        await db.connect()

        userEmail = `user_${Date.now()}@mail.com`

        // create normal user
        await axios.post(`${BASE_URL}/register`, {
            name: "normal user",
            email: userEmail,
            password: "Password123",
        })

        const userLogin = await axios.post(`${BASE_URL}/login`, {
            email: userEmail,
            password: "Password123",
        })

        userToken = userLogin.data.access_token

        // fetch test user uuid
        const result = await db.query(`SELECT id FROM users WHERE email = $1`, [userEmail])
        testUserId = result.rows[0].id

        // admin login
        const adminLogin = await axios.post(`${BASE_URL}/login`, {
            email: "admin@example.com",
            password: "Password123",
        })

        adminToken = adminLogin.data.access_token

        // fetch admin user uuid
        const adminResult = await db.query(`SELECT id FROM users WHERE email = $1`, ["admin@example.com"])
        adminUserId = adminResult.rows[0].id

    })

    // test that normal user cannot access all users
    it("normal user cannot access all users", async () => {
        try { await axios.get(`${BASE_URL}/users`, {
            headers: { Authorization: userToken }
        }) }

        catch (err) {
            expect(err.response.status).toBe(403)
        }
    })

    // test that admin can access all users
    it("admin can access all users", async () => {
        const res = await axios.get(`${BASE_URL}/users`, {
            headers: { Authorization: adminToken }
        })
        expect(res.status).toBe(200)
    })

    // test that normal user can update own profile
    it("normal user can update own profile", async () => {
        const res = await axios.put(`${BASE_URL}/users/me`, {
            name: "Updated User",
            email: userEmail,
            password: "Password123",
        }, {
            headers: { Authorization: userToken }
        })
        expect(res.status).toBe(200)
    })

    // test that normal user cannot update other user's profile
    it("normal user cannot update other user's profile", async () => {
        try { await axios.put(`${BASE_URL}/users/${adminUserId}`, {
            name: "Hacked User",
            email: "hacked@test.com",
            password: "Password123",
        }, {
            headers: { Authorization: userToken }
        }) }

        catch (err) {
            expect(err.response.status).toBe(403)
        }
    })

    // test that admin can update any user profile
    it("admin can update any user's profile", async () => {
        const res = await axios.put(`${BASE_URL}/users/${testUserId}`, {
            name: "Admin Updated User",
            email: "admin_updated@test.com",
            password: "Password123",
        }, {
            headers: { Authorization: adminToken }
        })
        expect(res.status).toBe(200)
    })

    // normal user cannot delete any user
    it("normal user cannot delete any user", async () => {
        try { await axios.delete(`${BASE_URL}/users/${adminUserId}`, {
            headers: { Authorization: userToken }
        }) }

        catch (err) {
            expect(err.response.status).toBe(403)
        }
    })

    // admin can delete any user
    it("admin can delete any user", async () => {
        const res = await axios.delete(`${BASE_URL}/users/${testUserId}`, {
            headers: { Authorization: adminToken }
        })
        expect(res.status).toBe(200)
    })

    // delete the test user and close db connection
    afterAll(async () => {
        await db.query(`DELETE FROM users WHERE email = $1`, [userEmail])
        await db.end()
    })
})

