import { useNavigate } from 'react-router'
import { Search } from 'lucide-react'

export default function NotFound() {
    const navigate = useNavigate()

    return (
        <div className="min-h-screen flex flex-col items-center justify-center text-center px-4 bg-white">
            <div className="w-20 h-20 rounded-full bg-[#EDE9F8] flex items-center justify-center mb-6">
                <Search className="w-9 h-9 text-[#8571C1]" />
            </div>
            <h1 className="text-6xl font-semibold text-[#8571C1] mb-2">404</h1>
            <h2 className="text-xl font-semibold text-gray-800 mb-2">Page not found</h2>
            <p className="text-sm text-gray-500 max-w-xs leading-relaxed mb-8">
                The page you're looking for doesn't exist or has been moved.
                Double-check the URL or head back home.
            </p>
            <div className="">
                <button
                    onClick={() => navigate('/dashboard')}
                    className="px-6 py-2.5 bg-[#8571C1] text-white rounded-xl text-sm font-medium hover:bg-[#7260b3] transition-colors"
                >
                    Back to dashboard
                </button>
            </div>
        </div>
    )
}