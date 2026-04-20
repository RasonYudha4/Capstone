import { useNavigate } from 'react-router'
import { LockKeyhole } from 'lucide-react'

export default function Unauthorized() {
    const navigate = useNavigate()

    return (
        <div className="min-h-screen flex flex-col items-center justify-center text-center px-4 bg-white">
            <div className="w-20 h-20 rounded-full bg-[#EDE9F8] flex items-center justify-center mb-6">
                <LockKeyhole className="w-9 h-9 text-[#8571C1]" />
            </div>
            <h1 className="text-6xl font-semibold text-[#8571C1] mb-2">401</h1>
            <h2 className="text-xl font-semibold text-gray-800 mb-2">Access denied</h2>
            <p className="text-sm text-gray-500 max-w-xs leading-relaxed mb-8">
                You don't have permission to view this page. Contact your
                administrator or sign in with a different account.
            </p>
            <div className="flex gap-3">
                <button
                    onClick={() => navigate('/login')}
                    className="px-6 py-2.5 bg-[#8571C1] text-white rounded-xl text-sm font-medium hover:bg-[#7260b3] transition-colors"
                >
                    Sign in
                </button>
                <button
                    onClick={() => navigate(-1)}
                    className="px-6 py-2.5 border border-[#8571C1] text-[#8571C1] rounded-xl text-sm font-medium hover:bg-[#EDE9F8] transition-colors"
                >
                    Go back
                </button>
            </div>
        </div>
    )
}