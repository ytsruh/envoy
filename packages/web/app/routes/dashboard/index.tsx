import { Button } from "~/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "~/components/ui/card";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "~/components/ui/tabs";
import { FileDown, PlusCircle } from "lucide-react";
import { projects } from "~/lib/data";
import { VariableTable } from "~/components/dashboard/variable-table";

export default function DashboardPage() {
  const project = projects[0]; // For demo purposes, we'll use the first project

  return (
    <>
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold md:text-2xl font-headline">{project.name}</h1>
      </div>
      <Tabs defaultValue="development" className="w-full">
        <div className="flex items-center">
          <TabsList>
            {project.environments.map((env) => (
              <TabsTrigger key={env.id} value={env.id}>
                {env.name}
              </TabsTrigger>
            ))}
          </TabsList>
          <div className="ml-auto flex items-center gap-2">
            <Button size="sm" variant="outline">
              <FileDown className="h-4 w-4 mr-2" />
              Export
            </Button>
            <Button size="sm">
              <PlusCircle className="h-4 w-4 mr-2" />
              Add Variable
            </Button>
          </div>
        </div>
        {project.environments.map((env) => (
          <TabsContent key={env.id} value={env.id}>
            <Card>
              <CardHeader>
                <CardTitle>{env.name} Variables</CardTitle>
                <CardDescription>
                  Manage environment variables for your {env.name.toLowerCase()} environment.
                </CardDescription>
              </CardHeader>
              <CardContent>
                <VariableTable variables={env.variables} />
              </CardContent>
            </Card>
          </TabsContent>
        ))}
      </Tabs>
    </>
  );
}
