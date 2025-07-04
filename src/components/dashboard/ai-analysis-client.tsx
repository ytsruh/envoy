"use client"

import { useState } from "react"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { BrainCircuit, Loader2, AlertTriangle, CheckCircle2 } from "lucide-react"
import { analyzeConfiguration, AnalyzeConfigurationInput, AnalyzeConfigurationOutput } from "@/ai/flows/analyze-configuration"
import type { Project } from "@/lib/data"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"

export default function AiAnalysisClient({ project }: { project: Project }) {
  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [result, setResult] = useState<AnalyzeConfigurationOutput | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [selectedEnv, setSelectedEnv] = useState<string>(project.environments[0].id)

  const handleAnalyze = async () => {
    setIsLoading(true);
    setError(null);
    setResult(null);

    const environment = project.environments.find(e => e.id === selectedEnv)
    if (!environment) {
        setError("Selected environment not found.");
        setIsLoading(false);
        return;
    }

    const variables = environment.variables.reduce((acc, v) => {
        acc[v.key] = v.value;
        return acc;
    }, {} as Record<string, string>);

    const input: AnalyzeConfigurationInput = {
      projectName: project.name,
      environmentName: environment.name,
      environmentVariables: variables
    }

    try {
      const analysisResult = await analyzeConfiguration(input);
      setResult(analysisResult);
    } catch (e) {
      setError("An error occurred during analysis.");
      console.error(e);
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <Dialog open={isOpen} onOpenChange={(open) => {
        setIsOpen(open);
        if(!open) {
            setResult(null);
            setError(null);
        }
    }}>
      <DialogTrigger asChild>
        <Button size="sm" variant="outline" className="bg-accent/10 border-accent/50 text-accent-foreground hover:bg-accent/20">
          <BrainCircuit className="h-4 w-4 mr-2" />
          AI Analysis
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[625px]">
        <DialogHeader>
          <DialogTitle className="font-headline flex items-center gap-2"><BrainCircuit /> AI Configuration Insights</DialogTitle>
          <DialogDescription>
            Select an environment to analyze its variables for potential issues.
          </DialogDescription>
        </DialogHeader>
        <div className="grid gap-4 py-4">
            <div className="flex items-center gap-4">
                <Select value={selectedEnv} onValueChange={setSelectedEnv}>
                    <SelectTrigger className="w-full">
                        <SelectValue placeholder="Select environment" />
                    </SelectTrigger>
                    <SelectContent>
                        {project.environments.map(env => (
                            <SelectItem key={env.id} value={env.id}>{env.name}</SelectItem>
                        ))}
                    </SelectContent>
                </Select>
                <Button onClick={handleAnalyze} disabled={isLoading} className="w-[150px]">
                    {isLoading ? <Loader2 className="h-4 w-4 animate-spin" /> : "Analyze"}
                </Button>
            </div>

            {error && (
                <Alert variant="destructive">
                    <AlertTriangle className="h-4 w-4" />
                    <AlertTitle>Error</AlertTitle>
                    <AlertDescription>{error}</AlertDescription>
                </Alert>
            )}

            {result && (
                 <div className="space-y-4 pt-4">
                    <Alert variant={result.alerts.length > 0 ? "destructive" : "default"} className={result.alerts.length === 0 ? "border-green-500/50" : ""}>
                       {result.alerts.length > 0 ? <AlertTriangle className="h-4 w-4" /> : <CheckCircle2 className="h-4 w-4 text-green-500" />}
                      <AlertTitle>{result.alerts.length > 0 ? `${result.alerts.length} Alert(s) Found` : "No Alerts Found"}</AlertTitle>
                      <AlertDescription>
                          {result.summary}
                      </AlertDescription>
                    </Alert>

                    {result.alerts.length > 0 && (
                        <div className="prose prose-sm dark:prose-invert max-w-none rounded-md border p-4 max-h-60 overflow-auto">
                            <h4 className="font-semibold mt-0">Details:</h4>
                            <ul className="pl-5 my-2">
                                {result.alerts.map((alert, index) => (
                                    <li key={index} className="mb-1">{alert}</li>
                                ))}
                            </ul>
                        </div>
                    )}
                 </div>
            )}
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setIsOpen(false)}>Close</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
