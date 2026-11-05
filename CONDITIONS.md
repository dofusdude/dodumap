# Unity condition interpretation

`ParseConditionUnity` owns parsing and semantic interpretation for items and combat entities. It preserves parentheses, case-sensitive codes, negative values, compound arguments and equipment operators. Published conditions describe client-side constraints, not the complete server eligibility formula. Confirmed server-owned predicates are excluded during mapping, never by the API renderer.

## Evidence

- DofusDB readable criteria: `https://api.dofusdb.fr/criterion/<encoded-expression>?lang=en`.
- Client factory and individual criterion classes: https://github.com/bot4dofus/Datafus/tree/master/data/A/DofusInvoker/scripts/com/ankamagames/dofus/datacenter/items/criterion
- Current Unity dump and decompiled factory: https://github.com/leonardo-spy/Jondo.Unity.WebClient/tree/master/Decompiled (inspected 2026-11-02). `StaticCriterionCriterion.omh` and `AlwaysTrueCriterion.omh` share RVA `0xDBDD50`. The `mw` factory maps static predicates to that class; it has no gameplay handler for the server-variable, raid-score, lowercase alteration and other server-owned predicates excluded by `serverOnlyCondition`.

Examples checked against DofusDB:

- `DD<30,21172`: item 21172 received less than 30 days ago. `DH` uses hours.
- `Pj>41,9`: Hunter profession level greater than 9.
- `POE8334` / `POX8334`: equipped / not equipped with Father Kwismas Hat.
- `SC!5`: not on a Temporis-type server.
- `ST=2`: server season 2 active.
- `HA=50`: Minor Uchronic Elixir alteration active.
- `Ms=0`: outside an arena; `Mt=1`: in a dungeon; `Az=1`: area under anomaly.
- `Ss!622,2970`: Fuji Snowfoux not present in Snowfoux Caverns.

## Server-only boundary

Do not invent the contents of server predicates. DofusDB returns no readable criteria for sampled `Sc`, `Sv`, `Pc`, `PX`, `RV`, and `ha` rules. These are not exposed as gameplay conditions. Static predicates always pass locally: `A & Sc` reduces to `A`, while `A | Sc` is locally unrestricted. Parentheses are preserved and simplification happens in dodumap.

Unclassified future codes are not automatically labeled server-only. They remain diagnostic mapping failures for the coverage audit to detect. No runtime API decision hides a condition tree. Missing display names for historical entity IDs do not change a verified predicate's meaning: typed references retain the original ID, and the API omits a missing name rather than inventing one.

Quest-objective operators require semantic conversion: `=`/`!` test active/not active, while `>`/`<` test completed/not completed. The API uses the corresponding semantic kind and equality/inequality, not a numeric comparison of objective IDs. Uppercase characteristic criteria describe total characteristics; lowercase versions describe additional characteristics.

Entity-reference catalogs come from doduda's downloaded raw data; no network lookup is performed during mapping or API requests.
