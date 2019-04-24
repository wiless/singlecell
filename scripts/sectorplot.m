close all

cfg=loadjson('OutputSetting.json');

ISD=cfg.ISD;


table=load('table700MHz.dat');
load uelocations.dat 
load bslocations.dat
% ISD=3200*2;
if ~exist('ISD') 
    disp('Set the value of ISD variable')
    return
end

% 
 load antennalocations.dat
 antennalocations=antennalocations(:,2:end);


%x=fileread('antennaArray.json');
%antennas=JSON.parse(x);
%antennalocations=[];
 
%for k=1:length(antennas)
%antennalocations(k,:)=[antennas{k}.SettingAAS.Centre.X antennas{k}.SettingAAS.Centre.Y];
%end

stable=sortrows(table,1);
rows=size(stable,1);
stable=[stable uelocations(1:rows,2) uelocations(1:rows,3) angle(uelocations(1:rows,2)+i*uelocations(1:rows,3))*180/pi];
 

%uepos=complex(uelocations(:,2),uelocations(:,3));
%DISTCOL=size(stable,2);
%for r=1:rows
%	bsloc=complex(bslocations(stable(r,7)+1,2),bslocations(stable(r,7)+1,3));
%    distance(r,1)=abs(uepos(r)-bsloc); 
% end
% stable(:,DISTCOL+1)=distance;


% CASE A1,B1,B2
% for SNR
%nodecolortable=[uelocations(:,1) stable(:,2) uelocations(:,2:3) stable(:,6)-stable(:,4) ]; % getting SINR  else stable(:,8)
% for SINR
nodecolortable=[uelocations(:,1) stable(:,2) uelocations(:,2:3) stable(:,8) stable(:,6) ]; % getting SINR  else stable(:,8)

% CASE A2
% FILTER UE beyond 1000m from center
%findx=find(distance>0); 
%stable=stable(findx,:);
%nodecolortable=[uelocations(findx,1) stable(:,2) uelocations(findx,2:3) stable(:,8)];



% figure
% sp=plot(stable(:,10),stable(:,11),'m.')
% for k=1:(length(bslocations)/3)
% plot(uelocations([1:400]+400*(k-1),2),uelocations([1:400]+400*(k-1),3),'.');
% hold all
% end


% syssinr=stable(uelocations(:,6)>3500,8);
syssinr=stable( : ,8);

% % FILTER positive SINR users only
    % stable=stable(find(stable(:,8)>-3),:);
      %stable=stable(find(stable(:,8)<=10),:)

figure(1) 
cdfplot(table(:,6));title('Coupling Loss');
figure(10)
cdfplot(syssinr);title('Geometric SINR');

figure(2)
[Nrows Ncols]=size(stable);
NUEsPerCell=cfg.NumUEperCell;
cell=3;
uerows=[1:NUEsPerCell]+NUEsPerCell*(cell-1);
stable(Nrows,Ncols+3)=0;
for indx=1:Nrows
    findx=find(uelocations(:,1)==stable(indx,1));
    extracols=[uelocations(findx,2:3) radtodeg(angle(uelocations(findx,2)+i*uelocations(findx,3)))];
    
    stable(indx,Ncols+1:Ncols+3)=extracols;
end
hold on;

plot(bslocations(:,2),bslocations(:,3),'*k','MarkerSize',10)
hold on;
plot(antennalocations(:,1),antennalocations(:,2),'Or','MarkerSize',10) 

% stable=stable(1:500,:);
bestbsid=stable(:,7);
sinrs=stable(:,8);
distances=stable(:,9);
activbs=find(bslocations(:,7)==1);
drawPolyGon(complex(bslocations(activbs,2),bslocations(activbs,3)),ISD/sqrt(3),'k');
inactivbs=find(bslocations(:,7)==0);
drawPolyGon(complex(bslocations(inactivbs,2),bslocations(inactivbs,3)),ISD/sqrt(3),'.');
% drawPolyGon(complex(antennalocations(:,1),antennalocations(:,2)),ISD/sqrt(3),'b',2);
%  
% clust1=20:37;
% drawPolyGon(complex(bslocations(clust1,2),bslocations(clust1,3)),ISD/2,'g');
% drawPolyGon(complex(antennalocations(clust1,1),antennalocations(clust1,2)),ISD/2,'g');

nSectors=3; 
nCells=size(bslocations,1)/nSectors;
MAXSINR=90
hold on;
k=0:nCells-1;
directions=[0 120 -120];
scolors=['r*' 'k*' 'b*'];

sec0bs=bslocations(find(bslocations(:,6)==0),1);
sec1bs=bslocations(find(bslocations(:,6)==120),1);
sec2bs=bslocations(find(bslocations(:,6)==-120),1);

selectedUEs0=stable(find((bestbsid>=min(sec0bs)).*(bestbsid<=max(sec0bs))),1);
selectedUEs1=stable(find((bestbsid>=min(sec1bs)).*(bestbsid<=max(sec1bs))),1);
selectedUEs2=stable(find((bestbsid>=min(sec2bs)).*(bestbsid<=max(sec2bs))),1);

sec0pos=FindUEs(selectedUEs0, uelocations,1);
sec1pos=FindUEs(selectedUEs1, uelocations,1);
sec2pos=FindUEs(selectedUEs2, uelocations,1);
 
h=plot(uelocations(sec0pos,2),uelocations(sec0pos,3),'r*');hold on
 plot(uelocations(sec1pos,2),uelocations(sec1pos,3),'k*');
 plot(uelocations(sec2pos,2),uelocations(sec2pos,3),'b*');
 legend ('sec0','sec1','sec2')


grid on

nodecolorplot


figure(1)
fname=sprintf('SINR_%.0f_%.0f',cfg.ISD,cfg.AntennaVTilt);
saveas(gcf,fname,'fig')
saveas(gcf,fname,'jpg')

%figure(2)
%saveas(gcf,'AssociatedSector','fig')
%saveas(gcf,'AssociatedSector','jpg')

figure(3)
saveas(gcf,'SingleCellCoverage','fig')
%saveas(gcf,'SingleCellCoverage','jpg')




