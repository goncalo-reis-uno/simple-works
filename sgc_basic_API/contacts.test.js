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

describe("Tests to verify contact routes", () => {
  let userToken
  let adminToken

  let userEmail

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

    // admin login
    const adminLogin = await axios.post(`${BASE_URL}/login`, {
      email: "admin@example.com",
      password: "Password123",
    })

    adminToken = adminLogin.data.access_token
  })


  // create 2 contacts for normal user
  it("user creates 2 contacts", async () => {
    const c1 = await axios.post(
      `${BASE_URL}/contacts`,
      {
        name: "User Contact 1",
        email: "u1@test.com",
        phone: "111111111",
        address: "Porto 1",
      },
      { headers: { Authorization: userToken } }
    )

    const c2 = await axios.post(
      `${BASE_URL}/contacts`,
      {
        name: "User Contact 2",
        email: "u2@test.com",
        phone: "222222222",
        address: "Porto 2",
      },
      { headers: { Authorization: userToken } }
    )

    expect(c1.status).toBe(201)
    expect(c2.status).toBe(201)
  })

  // create 2 contacts for admin
  it("admin creates 2 contacts", async () => {
    const c1 = await axios.post(
      `${BASE_URL}/contacts`,
      {
        name: "Admin Contact 1",
        email: "a1@test.com",
        phone: "333333333",
        address: "Lisbon 1",
      },
      { headers: { Authorization: adminToken } }
    )

    const c2 = await axios.post(
      `${BASE_URL}/contacts`,
      {
        name: "Admin Contact 2",
        email: "a2@test.com",
        phone: "444444444",
        address: "Lisbon 2",
      },
      { headers: { Authorization: adminToken } }
    )

    expect(c1.status).toBe(201)
    expect(c2.status).toBe(201)
  })


  // see all contacts (normal user)
  it("user should only see their own contacts", async () => {
    const res = await axios.get(`${BASE_URL}/contacts`, {
      headers: { Authorization: userToken },
    })

    expect(res.status).toBe(200)

    // user should only have 2 contacts
    expect(res.data.length).toBe(2)

    // ensure no admin contacts leak
    const hasAdminContact = res.data.some((c) =>
      c.name.includes("Admin")
    )

    expect(hasAdminContact).toBe(false)
  })

  // see all contacts (admin)
  it("admin should see all contacts", async () => {
    const res = await axios.get(`${BASE_URL}/contacts`, {
      headers: { Authorization: adminToken },
    })

    expect(res.status).toBe(200)

    expect(res.data.length).toBe(4)
  })

  // filtering
  it("should filter contacts by name", async () => {
    const res = await axios.get(
      `${BASE_URL}/contacts?name=User Contact`,
      { headers: { Authorization: userToken } }
    )

    expect(res.status).toBe(200)
    expect(res.data.length).toBe(2)
  })

  // sorting
  it("should sort contacts", async () => {
    const res = await axios.get(
      `${BASE_URL}/contacts?sort=created_at&order=desc`,
      { headers: { Authorization: adminToken } }
    )

    expect(res.status).toBe(200)
    expect(Array.isArray(res.data)).toBe(true)
  })

  // get contact by id
  it("user can get their own contact by id", async () => {
    const list = await axios.get(`${BASE_URL}/contacts`, {
      headers: { Authorization: userToken },
    })
    const contactId = list.data[0].id

    const res = await axios.get(`${BASE_URL}/contacts/${contactId}`, {
      headers: { Authorization: userToken },
    })

    expect(res.status).toBe(200)
    expect(res.data.id).toBe(contactId)
  })

  it("user cannot get admin contact by id", async () => {
    const list = await axios.get(`${BASE_URL}/contacts`, {
      headers: { Authorization: adminToken },
    })
    const adminContact = list.data.find((c) => c.name.includes("Admin"))

    try {
      await axios.get(`${BASE_URL}/contacts/${adminContact.id}`, {
        headers: { Authorization: userToken },
      })
      expect(true).toBe(false) // shouldnt reach here
    } catch (e) {
      expect(e.response.status).toBe(404)
    }
  })

  // update contact
  it("user can update their own contact", async () => {
    const list = await axios.get(`${BASE_URL}/contacts`, {
      headers: { Authorization: userToken },
    })
    const contactId = list.data[0].id

    const res = await axios.put(
      `${BASE_URL}/contacts/${contactId}`,
      {
        name: "Updated Contact",
        email: "updated@test.com",
        phone: "999999999",
        address: "Updated Address",
      },
      { headers: { Authorization: userToken } }
    )

    expect(res.status).toBe(200)
  })

  it("user cannot update admin contact", async () => {
    const list = await axios.get(`${BASE_URL}/contacts`, {
      headers: { Authorization: adminToken },
    })
    const adminContact = list.data.find((c) => c.name.includes("Admin"))

    try {
      await axios.put(
        `${BASE_URL}/contacts/${adminContact.id}`,
        {
          name: "Hacked",
          email: "hacked@test.com",
          phone: "000000000",
          address: "Hacked",
        },
        { headers: { Authorization: userToken } }
      )
      expect(true).toBe(false) // makes it fail on purpose
    } catch (e) {
      expect(e.response.status).toBe(404)
    }
  })

  // delete contact
  it("user cannot delete admin contact", async () => {
    const list = await axios.get(`${BASE_URL}/contacts`, {
      headers: { Authorization: adminToken },
    })
    const adminContact = list.data.find((c) => c.name.includes("Admin"))

    try {
      await axios.delete(`${BASE_URL}/contacts/${adminContact.id}`, {
        headers: { Authorization: userToken },
      })
      expect(true).toBe(false) // shouldnt reach here
    } catch (e) {
      expect(e.response.status).toBe(404)
    }
  })

  it("user can delete their own contact", async () => {
    const list = await axios.get(`${BASE_URL}/contacts`, {
      headers: { Authorization: userToken },
    })
    const contactId = list.data[0].id

    const res = await axios.delete(`${BASE_URL}/contacts/${contactId}`, {
      headers: { Authorization: userToken },
    })

    expect(res.status).toBe(200)
  })

  it("admin can delete any contact", async () => {
    const list = await axios.get(`${BASE_URL}/contacts`, {
      headers: { Authorization: adminToken },
    })
    const contact = list.data[0]

    const res = await axios.delete(`${BASE_URL}/contacts/${contact.id}`, {
      headers: { Authorization: adminToken },
    })

    expect(res.status).toBe(200)
  })


  // cleanup db
  afterAll(async () => {
    await db.query("DELETE FROM contacts WHERE email LIKE $1", [
      "%@test.com",
    ])

    await db.query("DELETE FROM users WHERE email = $1", [
      userEmail,
    ])

    await db.end()
  })
})