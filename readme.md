## Asana proxy for integration ide

### How to compile binary?
execute `make build`

### How to run app?
execute `make run`

### How to configure in phpstorm?
1. Go to Tools/Tasks/Servers
2. Add asana server
3. Fill fields in General tab: 
   - Server URL: http://localhost:8089
   - Username: Your asana user ID, example: 1201802048988821
   - Password: Your asana personal token
   - Project ID: asana project ID, example: 1201576078894109
   - User HTTP authorization: yes
4. Fill fields in Server configuration tab:
   - Task list URL: {serverUrl}/tasks?opt_fields=custom_fields,name&workspace=1198560554435486&assignee=me&opt_pretty=1&completed_since=now
   - Single task URL: {serverUrl}/tasks/{id}
   - | name          | value         |
     |---------------|---------------|
     | tasks         | data[*]       |
     | id            | id            |
     | summary       | name          |
     | singleTask-id | data.id       |

