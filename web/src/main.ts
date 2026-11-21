import { mount } from 'svelte'
import '@fontsource-variable/jetbrains-mono/wght.css'
import './app.css'
import App from './App.svelte'

mount(App, { target: document.getElementById('app')! })
