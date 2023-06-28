import { useEffect } from 'react'
import { useAuth, useUser } from "@clerk/clerk-react";

function AppSignedIn() {
    const { isLoaded, isSignedIn, userId, sessionId, getToken } = useAuth();
    const { isLoaded: isUserLoaded, isSignedIn: isUserSignedIn, user } = useUser();

    if (isLoaded && isSignedIn && isUserLoaded && isUserSignedIn) {
        useEffect(() => {
            getToken().then(tok => {
                let jwt = localStorage["clerk-db-jwt"];
                console.log("ignoring token: "+tok);

                let payload = {
                    userId: userId,
                    sessionId: sessionId,
                    token: jwt,
                    user: user,
                }
                const js = JSON.stringify(payload)
                const value = btoa(js);

                const timeout = setTimeout(() => {
                    // 👇️ redirects to an external URL
                    window.location.replace('http://localhost:3535/done?p='+value);
                }, 1000);

                return () => clearTimeout(timeout);
            })
        }, []);
    }

    return <>Redirecting...</>
}

export default AppSignedIn;
