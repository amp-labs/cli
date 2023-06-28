import './App.css'
import AppSignedIn from './signedin.tsx'
import { ClerkProvider, SignedIn, SignedOut, RedirectToSignIn } from "@clerk/clerk-react";

const clerkPubKey = "pk_test_bWlnaHR5LWtpbmdmaXNoLTY2LmNsZXJrLmFjY291bnRzLmRldiQ";

function App() {
  return (
    <ClerkProvider publishableKey={clerkPubKey}>
        <SignedIn>
            <AppSignedIn />
        </SignedIn>
        <SignedOut>
            <RedirectToSignIn />
        </SignedOut>
    </ClerkProvider>
  )
}

export default App
