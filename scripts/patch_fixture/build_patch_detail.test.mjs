import test from "node:test";
import assert from "node:assert/strict";

import { createAssetRegistry } from "./assets.mjs";
import { buildPatchDetail } from "./build_patch_detail.mjs";
import { heroLookupFromAssets } from "./lookups.mjs";

test("buildPatchDetail groups Doorman follow-up ability lines", async () => {
  const heroesLookup = heroLookupFromAssets([
    {
      id: 5,
      name: "The Doorman",
      images: {
        icon_image_small: "https://example.test/doorman.png",
      },
    },
  ]);

  const steamItem = {
    gid: "1826362059925616",
    title: "Gameplay Update - 03-06-2026",
    date: Date.UTC(2026, 2, 6) / 1000,
    contents: `[ General ]\n- Zipline speed increased\n[ Items ]\n- Active Reload: Cooldown reduced\n[ Heroes ]\n- Doorman\n- Gun now pierces through targets at 50% reduced damage\n- Call Bell time between charges increased from 4s to 6s\n- Doorway now has a timer icon above the ability\n- Luggage Cart is now 20% larger (20% wider hitbox as well)\n- Hotel Guest cast range increased from 6m to 7m`,
  };

  const fetchJson = async (url) => {
    if (!String(url).includes("/by-hero-id/5")) {
      throw new Error(`Unexpected URL: ${url}`);
    }
    return [
      { type: "ability", name: "Call Bell", image: "https://example.test/call-bell.png" },
      { type: "ability", name: "Doorway", image: "https://example.test/doorway.png" },
      { type: "ability", name: "Luggage Cart", image: "https://example.test/luggage-cart.png" },
      { type: "ability", name: "Hotel Guest", image: "https://example.test/hotel-guest.png" },
    ];
  };

  const { detail } = await buildPatchDetail({
    steamItem,
    allItems: [{ name: "Active Reload", type: "item", shop_image: "https://example.test/active-reload.png" }],
    assetsRegistry: createAssetRegistry(),
    fetchJson,
    heroesLookup,
  });

  const heroesSection = detail.sections.find((section) => section.kind === "heroes");
  assert.ok(heroesSection, "expected heroes section");

  const doorman = heroesSection.entries.find((entry) => entry.entityName === "Doorman");
  assert.ok(doorman, "expected Doorman entry");
  assert.equal(doorman.changes.length, 1);
  assert.equal(doorman.changes[0].text, "Gun now pierces through targets at 50% reduced damage");
  assert.deepEqual(
    doorman.groups.map((group) => group.title),
    ["Call Bell", "Doorway", "Luggage Cart", "Hotel Guest"],
  );
});
