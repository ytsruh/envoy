import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "~/components/ui/table";
import { MoreHorizontal, Eye } from "lucide-react";
import { Button } from "~/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/components/ui/dropdown-menu";
import type { Variable } from "~/lib/data";
import { cn } from "~/lib/utils";
import { useState } from "react";

interface VariableTableProps {
  variables: Variable[];
}

export function VariableTable({ variables }: VariableTableProps) {
  const [visibleVariables, setVisibleVariables] = useState<Record<string, boolean>>({});

  const toggleVisibility = (id: string) => {
    setVisibleVariables((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  if (variables.length === 0) {
    return (
      <div className="text-center text-muted-foreground py-12">
        <p>No variables in this environment.</p>
        <Button variant="link" className="mt-2">
          Add your first variable
        </Button>
      </div>
    );
  }

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Key</TableHead>
          <TableHead>Value</TableHead>
          <TableHead>Comment</TableHead>
          <TableHead className="w-[50px]">
            <span className="sr-only">Actions</span>
          </TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {variables.map((variable) => {
          const isVisible = visibleVariables[variable.id];

          return (
            <TableRow key={variable.id}>
              <TableCell className="font-medium font-code">{variable.key}</TableCell>
              <TableCell>
                <div className="flex items-center gap-2">
                  <span className={cn("font-code", !isVisible && "blur-xs select-none")}>
                    {isVisible ? variable.value : "********************"}
                  </span>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8"
                    onClick={() => toggleVisibility(variable.id)}>
                    <Eye className="h-4 w-4" />
                    <span className="sr-only">{isVisible ? "Hide" : "Show"} value</span>
                  </Button>
                </div>
              </TableCell>
              <TableCell className="text-muted-foreground">{variable.comment || "-"}</TableCell>
              <TableCell>
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button aria-haspopup="true" size="icon" variant="ghost">
                      <MoreHorizontal className="h-4 w-4" />
                      <span className="sr-only">Toggle menu</span>
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end">
                    <DropdownMenuItem>Edit</DropdownMenuItem>
                    <DropdownMenuItem>Copy Value</DropdownMenuItem>
                    <DropdownMenuItem className="text-destructive focus:text-destructive-foreground focus:bg-destructive">
                      Delete
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </TableCell>
            </TableRow>
          );
        })}
      </TableBody>
    </Table>
  );
}
