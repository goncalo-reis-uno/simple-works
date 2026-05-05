import {describe, it, expect} from "vitest";

const BaseURL = "http://localhost:8080";

describe("Route test", () => {
    it("should return current date", async () => {
        const response = await fetch(`${BaseURL}/date`);
        const data = await response.json();
        expect(response.status).toBe(200);
        expect(data).toHaveProperty("current_date");

        console.log(data.current_date);
        
        // Check format of date (DD/MM/YYYY) simple regex training
        expect(data.current_date).toMatch(/^\d{2}\/\d{2}\/\d{4}$/);
    });
});


