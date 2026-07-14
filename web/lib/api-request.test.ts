import { describe, expect, it } from "vitest";
import { API_REQUEST_OPTIONS } from "@/lib/api-request";

describe("API_REQUEST_OPTIONS", () => {
  it("keeps server-rendered API reads out of the filesystem-backed ISR cache", () => {
    expect(API_REQUEST_OPTIONS).toEqual({ cache: "no-store" });
  });
});
