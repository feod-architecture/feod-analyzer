# Profile Module

Owns user profile data and profile summary UI.

| Boundary | Details |
| --- | --- |
| Public API | profile model, `ProfileCard` |
| Consumers | profile page, billing, auth-aware widgets |

Profile consumers should use the module index to avoid coupling to model internals.

