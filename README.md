# Expense Tracker API

## Architecture

This project is structured with the Controller (Handler), Service, Repository pattern.
The Controller layer (or Handler in `net/http`) oversees the network logic, the Service layer handles all business/core logic, and finally the Repository layer performs data storage and access.

This should allow for a bigger degree of unit testing (with the use of Mocking the adjacent layers) and a more distinct separation of concerns between each layer.
