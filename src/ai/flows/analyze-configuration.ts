'use server';
/**
 * @fileOverview AI-powered tool analyzes environment variable configurations and alerts of suspicious or incorrect values.
 *
 * - analyzeConfiguration - A function that handles the environment variable configuration analysis process.
 * - AnalyzeConfigurationInput - The input type for the analyzeConfiguration function.
 * - AnalyzeConfigurationOutput - The return type for the analyzeConfiguration function.
 */

import {ai} from '@/ai/genkit';
import {z} from 'genkit';

const AnalyzeConfigurationInputSchema = z.object({
  environmentVariables: z.record(z.string()).describe('A key-value record of environment variables and their values.'),
  environmentName: z.string().describe('The name of the environment (e.g., development, staging, production).'),
  projectName: z.string().describe('The name of the project the environment belongs to.'),
});
export type AnalyzeConfigurationInput = z.infer<typeof AnalyzeConfigurationInputSchema>;

const AnalyzeConfigurationOutputSchema = z.object({
  alerts: z.array(z.string()).describe('An array of alerts for suspicious or incorrect environment variable values.'),
  summary: z.string().describe('A summary of the analysis results.'),
});
export type AnalyzeConfigurationOutput = z.infer<typeof AnalyzeConfigurationOutputSchema>;

export async function analyzeConfiguration(input: AnalyzeConfigurationInput): Promise<AnalyzeConfigurationOutput> {
  return analyzeConfigurationFlow(input);
}

const analyzeConfigurationPrompt = ai.definePrompt({
  name: 'analyzeConfigurationPrompt',
  input: {schema: AnalyzeConfigurationInputSchema},
  output: {schema: AnalyzeConfigurationOutputSchema},
  prompt: `You are an AI expert in analyzing environment variable configurations.

You will receive a set of environment variables, the environment name, and the project name.
Your task is to analyze the environment variables and identify any suspicious or incorrect values.

Consider the following factors when analyzing the environment variables:

- Variable names: Are there any unusual or potentially malicious variable names?
- Variable values: Are there any values that seem out of place or insecure (e.g., hardcoded passwords, API keys in development)?
- Environment: Are there any variables that are not appropriate for the given environment (e.g., debug settings in production)?
- Project: Are there any project-specific variables that are missing or incorrect?

Based on your analysis, generate a list of alerts and a summary of the results.

Environment Name: {{{environmentName}}}
Project Name: {{{projectName}}}
Environment Variables: {{JSON.stringify environmentVariables}}

Alerts:
Summary: `,
});

const analyzeConfigurationFlow = ai.defineFlow(
  {
    name: 'analyzeConfigurationFlow',
    inputSchema: AnalyzeConfigurationInputSchema,
    outputSchema: AnalyzeConfigurationOutputSchema,
  },
  async input => {
    const {output} = await analyzeConfigurationPrompt(input);
    return output!;
  }
);
