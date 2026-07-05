# GoPAMS
A lightweight, self-hosted, ITSM-inspired Password & Access Management
System built with **Go** . The project focuses on
secure password storage, organizational hierarchy, role-based access
control (RBAC), approval workflows, password lifecycle management, and
comprehensive audit logging.

----------------------------------------------------------------------------------
# Vision:

GoPAMS is designed as a single monolithic web application that is simple
enough to build from scratch while being architected for future
scalability.

Supported deployment scenarios:

-   Family Password Vault
-   Small Business Password Manager
-   Enterprise IT Password & Access Management

The architecture should remain identical regardless of deployment size.

*The Idea is not to be ServiceNow  but a reliable working product for simple use cases.*

-----------------------------------------------------------------------------------------

# Initial Setup :

❗Default Administrator Credentials :


Username: admin 

Password: AdminTempPassword123! 


Steps:
1. Ensure Docker Desktop is running.


2. Start the containers using :

"docker compose up --build -d"

3. Access the web interface at http://localhost:8080. The application will detect the first login and guide you through updating the password.

-----------------------------------------------------------------

# To delete previous data:

docker compose down -v 

docker compose up -d 

-------------------------------------------------------------------------

# Todo Implementation (read plan.md for details):

## First Phase:

* ~~Auth page~~
* ~~First time page prompt~~
### Refinement (planned-optionally):
* New password strength 
* Show password 

--------------------------------------------




