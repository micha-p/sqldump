# Layout (TODO)

sqlgopher will allow for different styles of presentation, depending on default settings, user agent or choice, and supplied cascaded style sheets. 
For security reasons, none of these layout options loads external libraries. 

Styles with navigation panel or custom layouts may need javascript.
The hardcoded templates rely on ASCII, the more sophisticated Document Object Model (DOM) allows for icons also.

### Query flags for using hardcoded templates (TODO)

?raw&			compact style with title, breadcrumb and menue (default)
?reduced&		reduced to main panel, best suited for cli-browsers
?navigation&	simple navigation panel on the left

### Query flags for using divisions / containers

-d  	Document Object Model (DOM)
-c file	supply customized style

