import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
  DialogFooter,
} from "~/components/ui/dialog";
import { Button } from "~/components/ui/button";
import { BrainCircuit } from "lucide-react";

import type { Project } from "~/lib/data";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "~/components/ui/select";

export default function AiAnalysisClient({ project }: { project: Project }) {
  const [isOpen, setIsOpen] = useState(false);

  const [selectedEnv, setSelectedEnv] = useState<string>(project.environments[0].id);

  return (
    <Dialog
      open={isOpen}
      onOpenChange={(open) => {
        setIsOpen(open);
        if (!open) {
          // Removed AI analysis related state resets
        }
      }}>
      <DialogTrigger asChild>
        <Button
          size="sm"
          variant="outline"
          className="bg-accent/10 border-accent/50 text-accent-foreground hover:bg-accent/20">
          <BrainCircuit className="h-4 w-4 mr-2" />
          AI Analysis
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-[625px]">
        <DialogHeader>
          <DialogTitle className="font-headline flex items-center gap-2">
            <BrainCircuit /> AI Configuration Insights
          </DialogTitle>
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
                {project.environments.map((env) => (
                  <SelectItem key={env.id} value={env.id}>
                    {env.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" onClick={() => setIsOpen(false)}>
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
