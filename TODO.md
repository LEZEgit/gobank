# Add a nextjs frontend to this

1. I want to protect routes R1 =  set of protected routes
2.
middleware:
when a client accesses these protected routes
    if client.Has(valid JWT token)
        show em the results
    else
        return error 
    

    client -> /protectedRoute -> auth_middleware -> valid JWT? -> Yes? -> run the actual handler
                                                        | No
                                                    return an error


    client -> /login -> assign a JWT (combine it with a unique field in account entity) and return it -> store the JWT on client side -> send that JWT with every request to protectedRoute