import { Button } from '@/components/ui/button';
import { Card } from '@/components/ui/card';
import { FileText, Users2, BookOpen, GitBranch } from 'lucide-react';
import Link from 'next/link';

export default function Home() {
  return (
    <div className="flex flex-col gap-8 p-8">
      <div className="flex flex-col gap-2">
        <h1 className="text-4xl font-bold">Welcome to SyncX</h1>
        <p className="text-lg text-muted-foreground">
          Your team's central hub for documentation and knowledge sharing
        </p>
      </div>

      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
        <Card className="p-6">
          <FileText className="h-12 w-12 text-primary" />
          <h2 className="mt-4 text-xl font-semibold">Documents</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Create and manage your team's documentation
          </p>
          <Button className="mt-4" asChild>
            <Link href="/documents">Browse Documents</Link>
          </Button>
        </Card>

        <Card className="p-6">
          <Users2 className="h-12 w-12 text-primary" />
          <h2 className="mt-4 text-xl font-semibold">Teams</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Collaborate with your team members
          </p>
          <Button className="mt-4" asChild>
            <Link href="/teams">View Teams</Link>
          </Button>
        </Card>

        <Card className="p-6">
          <BookOpen className="h-12 w-12 text-primary" />
          <h2 className="mt-4 text-xl font-semibold">Templates</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Start with pre-built document templates
          </p>
          <Button className="mt-4" asChild>
            <Link href="/templates">Browse Templates</Link>
          </Button>
        </Card>

        <Card className="p-6">
          <GitBranch className="h-12 w-12 text-primary" />
          <h2 className="mt-4 text-xl font-semibold">Version Control</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Track changes and manage versions
          </p>
          <Button className="mt-4" asChild>
            <Link href="/versions">View History</Link>
          </Button>
        </Card>
      </div>

      <div className="mt-8">
        <h2 className="text-2xl font-semibold mb-4">Recent Activity</h2>
        <div className="space-y-4">
          {/* Placeholder for recent activity - will be dynamic in the full implementation */}
          <Card className="p-4">
            <div className="flex items-center gap-4">
              <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
                <FileText className="h-5 w-5 text-primary" />
              </div>
              <div>
                <p className="font-medium">API Documentation Updated</p>
                <p className="text-sm text-muted-foreground">Updated by John Doe • 2 hours ago</p>
              </div>
            </div>
          </Card>
          <Card className="p-4">
            <div className="flex items-center gap-4">
              <div className="h-10 w-10 rounded-full bg-primary/10 flex items-center justify-center">
                <Users2 className="h-5 w-5 text-primary" />
              </div>
              <div>
                <p className="font-medium">New Team Member Added</p>
                <p className="text-sm text-muted-foreground">Jane Smith joined Engineering • 5 hours ago</p>
              </div>
            </div>
          </Card>
        </div>
      </div>
    </div>
  );
}