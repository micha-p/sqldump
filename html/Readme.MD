# Layout (TODO)

sqlgopher will allow for different styles of presentation, depending on default settings, user agent or choice, and eventually supplied cascaded style sheets. 
For security reasons, none of these layout options loads external libraries. Styles with navigation panel or custom layouts may need javascript.
The hardcoded templates rely on ASCII, the more sophisticated Document Object Model (DOM) allows for icons also.

### Options using hardcoded templates

-r		reduced to main panel, best suited for cli-browsers
-b		compact style with title, breadcrumb and menue (default)
-n		simple navigation panel on the left

### Options using divisions / containers

-d  	Document Object Model (DOM)
-c file	supply customized style

### Flags
-i		include information_schema, which is suppressed except in reduced or default style
-x		expert mode to access privileges, routines, triggers, views
