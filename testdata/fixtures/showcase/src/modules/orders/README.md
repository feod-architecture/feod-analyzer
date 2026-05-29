# Orders Module

Owns order records and order management tables.

| Boundary | Details |
| --- | --- |
| Public API | order model, `OrdersTable` |
| Consumers | reports, admin, checkout |

Order creation should enter through the module public API.

