import {describe, expect, it} from "vitest";

const BaseURL = 'http://localhost:8080';

const cases = [
    ["afonso@example.com", true],
    ["AfonSO123_@example.com", true],
    ["Af%ios@example-test.com", true], 
    ["FA+AC@example.com", true],
    ["teste-de-exemplo@example.com.pt", true],

    ["afonsoexample.com", false],
    ["afonso@examplecom", false],
    ["afonso@.com", false],
    ["afonso@com", false],
    ["afonso@exam_ple.com", false],
    ["afonso@exam%ple.com", false],
    ["afonso@exam+ple.com", false],
]

describe('Email validation', () => {
    cases.forEach(([email, expected]) => {
        it(`should return ${expected} for email: ${email}`, async () => {
            const encodedEmail = encodeURIComponent(email);
            const response = await fetch(`${BaseURL}/email?email=${encodedEmail}`);
            const data = await response.json();

            if (expected) {
                expect(data.valid).toBe(true);

            } else {
                expect(data.valid).toBe(false);
            }
        });
    });
});